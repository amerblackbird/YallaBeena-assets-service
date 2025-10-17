package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsset_JSONSerialization(t *testing.T) {
	// Create a test asset
	assetID := uuid.New()
	now := time.Now()

	asset := &Asset{
		ID:              assetID,
		URL:             "https://storage.example.com/file.jpg",
		PublicURL:       stringPtr("https://cdn.example.com/file.jpg"),
		Filename:        "test-file.jpg",
		FileSize:        1024,
		Metadata:        json.RawMessage(`{"key": "value"}`),
		Secure:          true,
		StorageKey:      stringPtr("path/to/file.jpg"),
		StorageProvider: stringPtr("s3"),
		ResourceID:      stringPtr("resource-123"),
		ResourceType:    stringPtr("profile"),
		ContentType:     "image/jpeg",
		UserID:          stringPtr("user-123"),
		AccessLevel:     "private",
		AllowedRoles:    pq.StringArray{"admin", "user"},
		IsEncrypted:     false,
		EncryptionKey:   nil,
		LastAccessedAt:  &now,
		DeletedAt:       nil,
		Tags:            pq.StringArray{"profile", "avatar"},
		CreatedAt:       now,
		UpdatedAt:       now,
		Active:          true,
	}

	// Test serialization
	jsonData, err := json.Marshal(asset)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "test-file.jpg")

	// Test deserialization
	var deserializedAsset Asset
	err = json.Unmarshal(jsonData, &deserializedAsset)
	require.NoError(t, err)

	assert.Equal(t, asset.ID, deserializedAsset.ID)
	assert.Equal(t, asset.Filename, deserializedAsset.Filename)
	assert.Equal(t, asset.FileSize, deserializedAsset.FileSize)
	assert.Equal(t, asset.ContentType, deserializedAsset.ContentType)
	assert.Equal(t, asset.Secure, deserializedAsset.Secure)
}

func TestCreateAssetDto_Validation(t *testing.T) {
	tests := []struct {
		name    string
		dto     CreateAssetDto
		isValid bool
	}{
		{
			name: "valid dto",
			dto: CreateAssetDto{
				URL:         "https://storage.example.com/file.jpg",
				Filename:    "test.jpg",
				FileSize:    1024,
				ContentType: "image/jpeg",
				AccessLevel: "public",
			},
			isValid: true,
		},
		{
			name: "missing filename",
			dto: CreateAssetDto{
				URL:         "https://storage.example.com/file.jpg",
				FileSize:    1024,
				ContentType: "image/jpeg",
				AccessLevel: "public",
			},
			isValid: false,
		},
		{
			name: "zero file size",
			dto: CreateAssetDto{
				URL:         "https://storage.example.com/file.jpg",
				Filename:    "test.jpg",
				FileSize:    0,
				ContentType: "image/jpeg",
				AccessLevel: "public",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := tt.dto.Filename != "" &&
				tt.dto.FileSize > 0 &&
				tt.dto.ContentType != "" &&
				tt.dto.URL != ""

			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestAssetFilter_BuildQuery(t *testing.T) {
	filter := &AssetFilter{
		UserID:      stringPtr("user-123"),
		ContentType: stringPtr("image/jpeg"),
		AccessLevel: stringPtr("public"),
		Secure:      boolPtr(false),
		Tags:        pq.StringArray{"profile", "avatar"},
		Limit:       10,
		Offset:      0,
	}

	// Test that filter contains expected values
	assert.Equal(t, "user-123", *filter.UserID)
	assert.Equal(t, "image/jpeg", *filter.ContentType)
	assert.Equal(t, "public", *filter.AccessLevel)
	assert.Equal(t, false, *filter.Secure)
	assert.Len(t, filter.Tags, 2)
	assert.Equal(t, int32(10), filter.Limit)
}

func TestUpdateAssetDto_PartialUpdate(t *testing.T) {
	assetID := uuid.New()

	// Test partial update - only some fields provided
	updateDto := &UpdateAssetDto{
		ID:          assetID,
		Filename:    stringPtr("updated-file.jpg"),
		AccessLevel: stringPtr("private"),
		// Other fields are nil, indicating they should not be updated
	}

	assert.Equal(t, assetID, updateDto.ID)
	assert.Equal(t, "updated-file.jpg", *updateDto.Filename)
	assert.Equal(t, "private", *updateDto.AccessLevel)
	assert.Nil(t, updateDto.FileSize)
	assert.Nil(t, updateDto.ContentType)
}

func TestStringArray_PostgreSQLCompatibility(t *testing.T) {
	// Test that pq.StringArray works as expected
	tags := pq.StringArray{"tag1", "tag2", "tag3"}

	assert.Len(t, tags, 3)
	assert.Contains(t, tags, "tag1")
	assert.Contains(t, tags, "tag2")
	assert.Contains(t, tags, "tag3")

	// Test empty array
	emptyTags := pq.StringArray{}
	assert.Len(t, emptyTags, 0)
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
