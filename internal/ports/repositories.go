package ports

import (
	"context"

	"assets-service/internal/core/domain"
)

// AssetsRepository defines the interface for asset data persistence
type AssetsRepository interface {
	CreateAsset(ctx context.Context, asset *domain.CreateAssetDto) (*domain.Asset, error)
	GetAssetByID(ctx context.Context, assetID string) (*domain.Asset, error)
	GetAssetsByUserID(ctx context.Context, userID string, limit, offset int32) ([]*domain.Asset, int32, error)
	UpdateAsset(ctx context.Context, asset *domain.UpdateAssetDto) (*domain.Asset, error)
	DeleteAsset(ctx context.Context, assetID string) error
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// LogActivity publishes user activity log event
	LogActivity(ctx context.Context, userID string, action string, metadata *domain.LogActivityMetadata) error

	PublishAvatarUpdated(ctx context.Context, userID string, avatarURL string) error

	// Stop stops publisher events
	Close() error
}

// EventConsumer defines the interface for consuming domain events
type EventConsumer interface {
	// Start starts consuming events
	Start(ctx context.Context) error

	// Stop stops consuming events
	Stop() error

	// RegisterHandler registers a handler for a specific event type
	RegisterHandler(eventType domain.EventType, handler EventHandler) error
}

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	Handle(ctx context.Context, event domain.DomainEvent) error
}

// KafkaTopics defines Kafka topic configuration
type KafkaTopics struct {
	ActivityLogEvents string `json:"activity.logs"`
}

// CacheService defines the interface for caching operations
type CacheService interface {
	// Set stores a value in cache
	Set(ctx context.Context, key string, value interface{}, ttl int) error

	// Get retrieves a value from cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Close closes the cache connection
	Close() error
}
