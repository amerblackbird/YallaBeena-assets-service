package kafka

import (
	"context"

	"assets-service/internal/core/domain"
	"assets-service/internal/ports"
)

// EventHandlers manages all event handlers
type EventHandlers struct {

	// Logger for logging events
	logger ports.Logger

	// Individual handlers
	assetsHandler *AssetsHandler
}

// NewEventHandlers creates a new event handlers manager
func NewEventHandlers(assetsRepo ports.AssetsRepository, logger ports.Logger) *EventHandlers {
	return &EventHandlers{
		logger:        logger,
		assetsHandler: NewActivityLogEventHandler(assetsRepo, logger),
	}
}

// RegisterHandlers registers all event handlers with the consumer
func (h *EventHandlers) RegisterHandlers(consumer ports.EventConsumer) error {

	// Register payment event handlers
	if err := consumer.RegisterHandler(domain.EventTypeLogActivity, h.assetsHandler); err != nil {
		return err
	}
	h.logger.Info("All event handlers registered successfully")
	return nil
}

type AssetsHandler struct {
	assetsRepo ports.AssetsRepository
	logger     ports.Logger
}

// NewActivityLogEventHandler creates a new activity log event handler
func NewActivityLogEventHandler(assetsRepo ports.AssetsRepository, logger ports.Logger) *AssetsHandler {
	return &AssetsHandler{
		assetsRepo: assetsRepo,
		logger:     logger,
	}
}

// Handle handles activity log events
func (h *AssetsHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	h.logger.Info("Handling activity log event",
		"event_type", string(event.Type),
		"event_id", event.ID,
		"aggregate_id", event.AggregateID)

	switch event.Type {
	case domain.EventTypeLogActivity:
		return h.handleCreateAttachmentProcessed(ctx, event)
	default:
		h.logger.Debug("Unhandled activity log event type", "event_type", string(event.Type))
		return nil
	}
}

func (h *AssetsHandler) handleCreateAttachmentProcessed(ctx context.Context, event domain.DomainEvent) error {
	// h.logger.Info("Processing activity log", "trip_id", event.AggregateID)
	// h.logger.Info("Event data", "data", event.Data)
	// // Marshal CreateActivityLogEvent
	// var createEvent events.ActivityLogEvent
	// data, err := json.Marshal(event.Data)
	// if err != nil {
	// 	h.logger.Error("Failed to marshal event data",
	// 		"event_type", string(event.Type),
	// 		"event_id", event.ID,
	// 		"aggregate_id", event.AggregateID,
	// 		"error", err)
	// 	return fmt.Errorf("failed to marshal event data: %w", err)
	// }
	// if err := json.Unmarshal(data, &createEvent); err != nil {
	// 	h.logger.Error("Failed to unmarshal event data",
	// 		"event_type", string(event.Type),
	// 		"event_id", event.ID,
	// 		"aggregate_id", event.AggregateID,
	// 		"error", err)
	// 	return fmt.Errorf("failed to unmarshal event data: %w", err)
	// }
	// activity, err := h.assetsRepo.CreateAsset(ctx, &domain.CreateAssetDto{
	// 	Filename:    createEvent.Filename,
	// 	ContentType: createEvent.ContentType,
	// 	FileSize:    createEvent.FileSize,
	// 	UserID:      createEvent.UserID,
	// 	Metadata:    createEvent.Metadata,
	// })
	// if err != nil {
	// 	h.logger.Error("Failed to create activity log",
	// 		"event_type", string(event.Type),
	// 		"event_id", event.ID,
	// 		"aggregate_id", event.AggregateID,
	// 		"error", err)
	// 	return fmt.Errorf("failed to create activity log: %w", err)
	// }

	// h.logger.Info("Activity log created successfully",
	// 	"event_type", string(event.Type),
	// 	"event_id", event.ID,
	// 	"aggregate_id", event.AggregateID,
	// 	"activity_id", activity.ID)

	return nil
}

func (h *AssetsHandler) handleActivityLogFailed(ctx context.Context, event domain.DomainEvent) error {
	h.logger.Info("Processing activity log failure", "trip_id", event.AggregateID)
	// Implementation for activity log failure handling
	return nil
}
