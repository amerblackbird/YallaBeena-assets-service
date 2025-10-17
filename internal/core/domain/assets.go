package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Asset represents an uploaded asset/file
type Asset struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	URL             string          `json:"url" db:"url"`                           // Storage URL
	PublicURL       string          `json:"public_url" db:"public_url"`             // Asset public URL if available
	Filename        string          `json:"filename" db:"filename"`                 // Original filename
	FileSize        int64           `json:"file_size" db:"file_size"`               // Size in bytes
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`                 // Additional metadata as JSON
	Secure          bool            `json:"secure" db:"secure"`                     // Whether the asset is stored securely
	StorageKey      *string         `json:"storage_key" db:"storage_key"`           // Key used in storage backend
	StorageProvider *string         `json:"storage_provider" db:"storage_provider"` // e.g., "s3", "gcs"
	ResourceID      *string         `json:"resource_id" db:"resource_id"`           // Associated resource ID
	ResourceType    *string         `json:"resource_type" db:"resource_type"`       // e.g., "profile_picture", "document"
	ContentType     string          `json:"content_type" db:"content_type"`         // MIME type
	UserID          *string         `json:"user_id" db:"user_id"`                   // ID of the user who uploaded the asset
	AccessLevel     string          `json:"access_level" db:"access_level"`         // e.g., "public", "private"
	AllowedRoles    pq.StringArray  `json:"allowed_roles" db:"allowed_roles"`       // Roles allowed to access
	IsEncrypted     bool            `json:"is_encrypted" db:"is_encrypted"`         // Whether the asset is encrypted
	EncryptionKey   *string         `json:"encryption_key" db:"encryption_key"`     // Key used for encryption if applicable
	LastAccessedAt  *time.Time      `json:"last_accessed_at" db:"last_accessed_at"` // Last accessed timestamp
	DeletedAt       *time.Time      `json:"deleted_at" db:"deleted_at"`             // Soft delete timestamp
	Tags            pq.StringArray  `json:"tags" db:"tags"`                         // Tags for categorization
	CreatedAt       string          `json:"created_at" db:"created_at"`             // Creation timestamp
	UpdatedAt       string          `json:"updated_at" db:"updated_at"`             // Last update timestamp
	Active          bool            `json:"active" db:"active"`                     // Whether the asset is active
	FileHash        string          `json:"file_hash" db:"file_hash"`               // SHA256 hash of the file for integrity
}

// CreateAssetDto represents the DTO for creating an asset
type CreateAssetDto struct {
	URL             string          `json:"url" db:"url"` // Asset
	PublicURL       *string         `json:"public_url" db:"public_url"`
	Filename        string          `json:"filename" db:"filename"`
	FileSize        int64           `json:"file_size" db:"file_size"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
	Secure          bool            `json:"secure" db:"secure"`
	FileHash        string          `json:"file_hash" db:"file_hash"`
	StorageKey      *string         `json:"storage_key" db:"storage_key"`
	StorageProvider *string         `json:"storage_provider" db:"storage_provider"`
	ResourceID      *string         `json:"resource_id" db:"resource_id"`
	ResourceType    *string         `json:"resource_type" db:"resource_type"`
	ContentType     string          `json:"content_type" db:"content_type"`
	UserID          *string         `json:"user_id" db:"user_id"`
	AccessLevel     string          `json:"access_level" db:"access_level"`
	AllowedRoles    pq.StringArray  `json:"allowed_roles" db:"allowed_roles"`
	IsEncrypted     bool            `json:"is_encrypted" db:"is_encrypted"`
	EncryptionKey   *string         `json:"encryption_key" db:"encryption_key"`
	Tags            pq.StringArray  `json:"tags" db:"tags"`
}

type UpdateAssetDto struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	URL             *string         `json:"url" db:"url"`
	PublicURL       *string         `json:"public_url" db:"public_url"`
	Filename        *string         `json:"filename" db:"filename"`
	FileSize        *int64          `json:"file_size" db:"file_size"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
	Secure          *bool           `json:"secure" db:"secure"`
	StorageKey      *string         `json:"storage_key" db:"storage_key"`
	StorageProvider *string         `json:"storage_provider" db:"storage_provider"`
	ResourceID      *string         `json:"resource_id" db:"resource_id"`
	ResourceType    *string         `json:"resource_type" db:"resource_type"`
	ContentType     *string         `json:"content_type" db:"content_type"`
	UserID          *string         `json:"user_id" db:"user_id"`
	AccessLevel     *string         `json:"access_level" db:"access_level"`
	AllowedRoles    pq.StringArray  `json:"allowed_roles" db:"allowed_roles"`
	IsEncrypted     *bool           `json:"is_encrypted" db:"is_encrypted"`
	EncryptionKey   *string         `json:"encryption_key" db:"encryption_key"`
	Tags            pq.StringArray  `json:"tags" db:"tags"`
	FileHash        string          `json:"file_hash" db:"file_hash"` // SHA256 hash of the file for integrity

}

// AssetFilter represents filters for querying assets
type AssetFilter struct {
	UserID          *string        `json:"user_id"`
	ContentType     *string        `json:"content_type"`
	ResourceType    *string        `json:"resource_type"`
	ResourceID      *string        `json:"resource_id"`
	AccessLevel     *string        `json:"access_level"`
	Secure          *bool          `json:"secure"`
	IsEncrypted     *bool          `json:"is_encrypted"`
	StorageProvider *string        `json:"storage_provider"`
	Tags            pq.StringArray `json:"tags"`
	Limit           int32          `json:"limit"`
	Offset          int32          `json:"offset"`
}
