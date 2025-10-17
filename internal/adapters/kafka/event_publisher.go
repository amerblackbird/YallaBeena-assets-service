package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	config "assets-service/configs"
	"assets-service/internal/core/domain"
	"assets-service/internal/core/events"
	"assets-service/internal/ports"
)

// EventPublisher implements the EventPublisher interface using Kafka
type EventPublisher struct {
	writers map[string]*kafka.Writer
	logger  ports.Logger
	config  config.KafkaConfig
}

// NewEventPublisher creates a new Kafka event publisher
func NewEventPublisher(config config.KafkaConfig, logger ports.Logger) ports.EventPublisher {
	writers := make(map[string]*kafka.Writer)

	// Create writers for each topic
	topics := []string{
		config.Topics.ActivityLogs,
	}

	for _, topic := range topics {
		writers[topic] = &kafka.Writer{
			Addr:         kafka.TCP(config.Brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 10 * time.Millisecond,
		}
	}

	return &EventPublisher{
		writers: writers,
		logger:  logger,
		config:  config,
	}
}

func (p *EventPublisher) LogActivity(ctx context.Context, userID string, action string, metadata *domain.LogActivityMetadata) error {

	var meta domain.LogActivityMetadata
	if metadata != nil {
		meta = domain.LogActivityMetadata{
			IP:       metadata.IP,
			Device:   metadata.Device,
			Location: metadata.Location,
		}
	} else {
		meta = domain.LogActivityMetadata{
			IP:       "assets-service",
			Device:   "server",
			Location: "unknown",
		}
	}

	metadataJSON, err := json.Marshal(meta)
	if err != nil {
		p.logger.Error("Failed to marshal metadata", zap.Error(err))
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	event := events.LogActivityEvent{
		ID:        generateEventID(),
		UserID:    userID,
		Action:    action,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  metadataJSON,
	}

	domainEvent := domain.DomainEvent{
		ID:          generateEventID(),
		Type:        domain.EventTypeLogActivity,
		AggregateID: userID,
		Version:     1,
		Data:        eventToMap(event),
		Metadata: domain.EventMetadata{
			Source:        "auth-service",
			CorrelationID: getCorrelationID(ctx),
		},
		Timestamp: time.Now(),
	}

	return p.publishEvent(ctx, p.config.Topics.ActivityLogs, domainEvent)
}

// publishEvent publishes a domain event to Kafka
func (p *EventPublisher) publishEvent(ctx context.Context, topic string, event domain.DomainEvent) error {
	writer, exists := p.writers[topic]
	if !exists {
		return fmt.Errorf("no writer found for topic: %s", topic)
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal event",
			zap.String("event_type", string(event.Type)),
			zap.String("aggregate_id", event.AggregateID),
			zap.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: eventJSON,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(string(event.Type))},
			{Key: "event-id", Value: []byte(event.ID)},
			{Key: "correlation-id", Value: []byte(event.Metadata.CorrelationID)},
		},
	}

	err = writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.Error("Failed to publish event",
			zap.String("topic", topic),
			zap.String("event_type", string(event.Type)),
			zap.String("aggregate_id", event.AggregateID),
			zap.Error(err))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info("Event published successfully",
		"topic", topic,
		"event_type", string(event.Type),
		"event_id", event.ID,
		"aggregate_id", event.AggregateID)

	return nil
}

// Close closes all Kafka writers
func (p *EventPublisher) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Failed to close writer",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}
	return nil
}
