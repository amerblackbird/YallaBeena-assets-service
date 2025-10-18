package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	config "assets-service/configs"
	"assets-service/internal/core/domain"
	"assets-service/internal/ports"
)

// EventConsumer implements the EventConsumer interface using Kafka
type EventConsumer struct {
	readers  map[string]*kafka.Reader
	handlers map[domain.EventType]ports.EventHandler
	logger   ports.Logger
	config   config.KafkaConfig
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewEventConsumer creates a new Kafka event consumer
func NewEventConsumer(config config.KafkaConfig, logger ports.Logger) ports.EventConsumer {
	readers := make(map[string]*kafka.Reader)

	// Create readers for topics we want to consume from
	consumeTopics := map[string]string{
		"users.events": config.Topics.UsersEvents,
	}

	for name, topic := range consumeTopics {
		readers[name] = kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.Brokers,
			Topic:       topic,
			GroupID:     config.GroupID,
			StartOffset: kafka.LastOffset,
			MinBytes:    10e3, // 10KB
			MaxBytes:    10e6, // 10MB
		})
	}

	return &EventConsumer{
		readers:  readers,
		handlers: make(map[domain.EventType]ports.EventHandler),
		logger:   logger,
		config:   config,
	}
}

// Start starts consuming events
func (c *EventConsumer) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Start a goroutine for each reader
	for name, reader := range c.readers {
		c.wg.Add(1)
		go c.consumeMessages(name, reader)
	}

	c.logger.Info("Event consumer started",
		zap.Int("readers", len(c.readers)),
		zap.String("group_id", c.config.GroupID))

	return nil
}

// Stop stops consuming events
func (c *EventConsumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}

	c.wg.Wait()

	// Close all readers
	for name, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.Error("Failed to close reader",
				zap.String("reader", name),
				zap.Error(err))
		}
	}

	c.logger.Info("Event consumer stopped")
	return nil
}

// RegisterHandler registers a handler for a specific event type
func (c *EventConsumer) RegisterHandler(eventType domain.EventType, handler ports.EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers[eventType] = handler
	c.logger.Info("Event handler registered",
		zap.String("event_type", string(eventType)))

	return nil
}

// consumeMessages consumes messages from a Kafka reader
func (c *EventConsumer) consumeMessages(readerName string, reader *kafka.Reader) {
	c.logger.Info("Starting message consumption",
		zap.String("reader", readerName))

	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Stopping message consumption",
				zap.String("reader", readerName))
			return
		default:
			message, err := reader.FetchMessage(c.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				c.logger.Error("Failed to fetch message",
					zap.String("reader", readerName),
					zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			if err := c.handleMessage(message); err != nil {
				c.logger.Error("Failed to handle message",
					zap.String("reader", readerName),
					zap.String("topic", message.Topic),
					zap.Int("partition", message.Partition),
					zap.Int64("offset", message.Offset),
					zap.Error(err))
			} else {
				// Commit the message only if handling was successful
				if err := reader.CommitMessages(c.ctx, message); err != nil {
					c.logger.Error("Failed to commit message",
						zap.String("reader", readerName),
						zap.Error(err))
				}
			}
		}
	}
}

// handleMessage handles a Kafka message
func (c *EventConsumer) handleMessage(message kafka.Message) error {

	// Log the received message
	c.logger.Info("Received message",
		zap.String("topic", message.Topic),
		zap.Int("partition", message.Partition),
		zap.Int64("offset", message.Offset),
		zap.Time("timestamp", message.Time))

	// Parse the domain event
	var domainEvent domain.DomainEvent
	if err := json.Unmarshal(message.Value, &domainEvent); err != nil {
		// Log the error and return
		c.logger.Error("Failed to unmarshal domain event",
			zap.String("topic", message.Topic),
			zap.Int("partition", message.Partition),
			zap.Int64("offset", message.Offset),
			zap.Error(err))
		return fmt.Errorf("failed to unmarshal domain event: %w", err)
	} else {
		// Log the parsed domain event
		c.logger.Info("Parsed domain event",
			zap.String("event_type", string(domainEvent.Type)),
			zap.String("event_id", domainEvent.ID),
			zap.String("aggregate_id", domainEvent.AggregateID),
			zap.Time("timestamp", domainEvent.Timestamp))
	}

	// Find and execute the handler
	c.mu.RLock()
	handler, exists := c.handlers[domainEvent.Type]
	c.mu.RUnlock()

	if !exists {
		c.logger.Info("No handler found for event type",
			zap.String("event_type", string(domainEvent.Type)))
		return nil // Not an error - we just don't handle this event type
	}

	c.logger.Info("Handling event",
		zap.String("event_type", string(domainEvent.Type)),
		zap.String("event_id", domainEvent.ID),
		zap.String("aggregate_id", domainEvent.AggregateID))

	// Add correlation ID to context
	ctx := context.WithValue(c.ctx, "correlation_id", domainEvent.Metadata.CorrelationID)

	// Handle the event
	if err := handler.Handle(ctx, domainEvent); err != nil {
		return fmt.Errorf("handler failed for event %s: %w", domainEvent.Type, err)
	}

	c.logger.Info("Event handled successfully",
		zap.String("event_type", string(domainEvent.Type)),
		zap.String("event_id", domainEvent.ID))

	return nil
}
