package kafka

import (
	"context"
	"encoding/json"

	"assets-service/internal/core/domain"
	"assets-service/internal/core/events"
	"assets-service/internal/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EventHandlers manages all event handlers
type EventHandlers struct {

	// Logger for logging events
	logger ports.Logger

	// Individual handlers
	assetsHandler *AssetsHandler
}

// NewEventHandlers creates a new event handlers manager
func NewEventHandlers(assetsService ports.AssetsService,
	eventPublisher ports.EventPublisher,
	logger ports.Logger) *EventHandlers {
	return &EventHandlers{
		assetsHandler: NewAssetEventHandler(assetsService, eventPublisher, logger),
		logger:        logger,
	}
}

// RegisterHandlers registers all event handlers with the consumer
func (h *EventHandlers) RegisterHandlers(consumer ports.EventConsumer) error {

	// Register payment event handlers
	if err := consumer.RegisterHandler(domain.EventTypeUserCreated, h.assetsHandler); err != nil {
		return err
	}
	h.logger.Info("All event handlers registered successfully")
	return nil
}

type AssetsHandler struct {
	assetsService  ports.AssetsService
	eventPublisher ports.EventPublisher
	logger         ports.Logger
}

// NewAssetEventHandler creates a new assets related event handler
func NewAssetEventHandler(assetsService ports.AssetsService, eventPublisher ports.EventPublisher, logger ports.Logger) *AssetsHandler {
	return &AssetsHandler{
		assetsService:  assetsService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle handles assets related events
func (h *AssetsHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	h.logger.Info("Handling events",
		"event_type", string(event.Type),
		"event_id", event.ID,
		"aggregate_id", event.AggregateID)

	switch event.Type {
	case domain.EventTypeUserCreated:
		err := h.handleUserCreated(ctx, event)
		if err != nil {
			// Todo:// Add task to retry or log failure
			h.logger.Error("Failed to handle user created event",
				"event_id", event.ID,
				"aggregate_id", event.AggregateID,
				"error", err)
			return err
		}
		h.logger.Info("User created event handled successfully",
			"event_id", event.ID,
			"aggregate_id", event.AggregateID)
		return nil
	default:
		h.logger.Debug("Unhandled activity log event type", "event_type", string(event.Type))
		return nil
	}
}

func (h *AssetsHandler) handleUserCreated(ctx context.Context, event domain.DomainEvent) error {
	h.logger.Info("Processing user created event", "user_id", event.AggregateID)

	var userCreatedEvent events.UserCreatedEvent
	data, err := json.Marshal(event.Data)
	if err != nil {
		h.logger.Error("Failed to marshal event data",
			"event_type", string(event.Type),
			"event_id", event.ID,
			"aggregate_id", event.AggregateID,
			"error", err)
		return err
	}
	if err := json.Unmarshal(data, &userCreatedEvent); err != nil {
		h.logger.Error("Failed to unmarshal event data",
			"event_type", string(event.Type),
			"event_id", event.ID,
			"aggregate_id", event.AggregateID,
			"error", err)
		return err
	}

	if userCreatedEvent.UserID == "" {
		h.logger.Error("User created event missing UserID",
			"event_type", string(event.Type),
			"event_id", event.ID,
			"aggregate_id", event.AggregateID)
		return nil // No UserID to process
	}

	// Create avatar asset for the new user (placeholder logic)

	// Process the user created event (e.g., create a welcome asset)

	h.logger.Info("User created event processed successfully",
		"user_id", userCreatedEvent.UserID)

	fileName := "avatar.png"
	resourceType := "users"
	contentType := "image/png"
	jsonMeta := json.RawMessage(`{"description":"Welcome avatar for new user"}`)

	fileData := domain.GetDefaultAvatar()

	fileSize := int64(len(fileData)) // Placeholder size

	createDto := &domain.CreateAssetDto{
		Filename:        fileName,
		ContentType:     contentType,
		FileSize:        fileSize,
		UserID:          &userCreatedEvent.UserID,
		Metadata:        jsonMeta,
		Secure:          false,
		Tags:            []string{},
		AccessLevel:     "public",
		AllowedRoles:    []string{},
		IsEncrypted:     false,
		EncryptionKey:   nil,
		StorageProvider: nil,
		ResourceID:      nil,
		ResourceType:    &resourceType,
	}

	// Call the service
	asset, err := h.assetsService.UploadAsset(ctx, createDto, fileData)
	if err != nil {
		h.logger.Error("Failed to upload asset", "error", err)
		return status.Errorf(codes.Internal, "failed to upload asset: %v", err)
	}

	// Publish event or log success
	err = h.eventPublisher.PublishAvatarUpdated(ctx, userCreatedEvent.UserID, asset.URL)
	if err != nil {
		h.logger.Error("Failed to publish avatar updated event", "error", err)
		return status.Errorf(codes.Internal, "failed to publish avatar updated event: %v", err)
	}

	h.logger.Info("Welcome avatar asset created successfully",
		"user_id", userCreatedEvent.UserID,
		"asset_id", asset.ID,
		"asset_url", asset.URL)

	return nil
}

func (h *AssetsHandler) handleActivityLogFailed(ctx context.Context, event domain.DomainEvent) error {
	h.logger.Info("Processing activity log failure", "trip_id", event.AggregateID)
	// Implementation for activity log failure handling
	return nil
}
