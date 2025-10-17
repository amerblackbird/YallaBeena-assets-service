package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"assets-service/internal/core/domain"
	"assets-service/internal/ports"
	utils "assets-service/internal/utils"

	"github.com/google/uuid"
)

// AssetsService implements the assets service interface
type AssetsService struct {
	assetsRepo     ports.AssetsRepository
	storageService ports.StoragesService
	cacheService   ports.CacheService
	logger         ports.Logger
}

// NewAssetsService creates a new assets service
func NewAssetsService(
	assetsRepo ports.AssetsRepository,
	storageService ports.StoragesService,
	cacheService ports.CacheService,
	logger ports.Logger) ports.AssetsService {
	return &AssetsService{
		assetsRepo:     assetsRepo,
		cacheService:   cacheService,
		storageService: storageService,
		logger:         logger,
	}
}

// UploadAsset uploads a new asset and returns metadata
func (s *AssetsService) UploadAsset(ctx context.Context, createDto *domain.CreateAssetDto, fileData []byte) (*domain.Asset, error) {
	s.logger.Info("Uploading asset", "filename", createDto.Filename, "user_id", createDto.UserID)

	// Generate unique asset ID
	assetID := uuid.New().String()

	// Generate unique slug for filename to avoid conflicts
	timestamp := time.Now().Unix()
	uniqueSlug := fmt.Sprintf("%d_%s", timestamp, createDto.Filename)

	// Generate file key for storage (handle null UserID)
	var fileKey string = ""
	if createDto.ResourceType != nil && *createDto.ResourceType != "" && createDto.ResourceID != nil && *createDto.ResourceID != "" {
		fileKey = fmt.Sprintf("%s/%s/%s", *createDto.ResourceType, *createDto.ResourceID, uniqueSlug)
	} else if createDto.ResourceType != nil && *createDto.ResourceType != "" && (createDto.ResourceID != nil || *createDto.ResourceID != "") {
		fileKey = fmt.Sprintf("%s/%s", *createDto.ResourceType, uniqueSlug)
	}
	// // Generate file key for storage (handle null UserID)
	// if createDto.UserID != nil && *createDto.UserID != "" {
	// 	fileKey = fmt.Sprintf("%s/%s", *createDto.UserID, fileKey)
	// }

	s.logger.Info("Generated file key for storage", "file_key", fileKey, "asset_id", assetID)

	// Upload file to storage
	assetURL, err := s.storageService.UploadFile(ctx, fileKey, fileData, createDto.ContentType)
	if err != nil {
		s.logger.Error("Failed to upload file to storage", "error", err, "file_key", fileKey)
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	s.logger.Info("File uploaded to storage", "file_key", fileKey, "asset_url", assetURL)

	// Add metadata including file hash for integrity
	fileHash := fmt.Sprintf("%x", sha256.Sum256(fileData))
	metadata := map[string]interface{}{
		"file_hash":        fileHash,
		"upload_timestamp": time.Now().Unix(),
		"storage_key":      fileKey,
	}
	if len(createDto.Metadata) > 0 {
		var customMetadata map[string]interface{}
		if err := json.Unmarshal(createDto.Metadata, &customMetadata); err == nil {
			for k, v := range customMetadata {
				metadata[k] = v
			}
		} else {
			s.logger.Warn("Failed to unmarshal custom metadata, ignoring", "error", err)
		}
	}

	// Convert metadata to JSON for storage
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, domain.NewDomainError(domain.UnableToMarshalError, "Failed to marshal metadata", err)
	}

	publicUrl := fmt.Sprintf("assets/%s", assetID)

	fileSize := int64(len(fileData))

	// Create asset DTO for repository
	assetDto := &domain.CreateAssetDto{
		StorageKey:      &fileKey,
		StorageProvider: utils.StringPtr("minio"),
		URL:             assetURL,
		PublicURL:       &publicUrl,
		Filename:        createDto.Filename,
		ContentType:     createDto.ContentType,
		FileSize:        fileSize,
		UserID:          createDto.UserID,
		Metadata:        metadataJSON,
		FileHash:        fileHash,
		Secure:          createDto.Secure,
		Tags:            createDto.Tags,
		AccessLevel:     createDto.AccessLevel,
		IsEncrypted:     createDto.IsEncrypted,
		ResourceID:      createDto.ResourceID,
		ResourceType:    createDto.ResourceType,
		EncryptionKey:   createDto.EncryptionKey,
	}

	// Save asset metadata to database
	savedAsset, err := s.assetsRepo.CreateAsset(ctx, assetDto)
	if err != nil {
		s.logger.Error("Failed to save asset metadata", "error", err, "asset_id", assetID)

		// Rollback: delete the file from storage if database save fails
		if deleteErr := s.storageService.DeleteFile(ctx, fileKey); deleteErr != nil {
			s.logger.Error("Failed to rollback file upload", "error", deleteErr, "file_key", fileKey)
		}
		return nil, domain.NewDomainError(domain.UnableToMarshalError, "Failed to save asset metadata", err)
	}

	// Cache the asset

	cacheKey := fmt.Sprintf("asset:%s", assetID)
	if err := s.cacheService.Set(ctx, cacheKey, savedAsset, 0); err != nil {
		s.logger.Error("Failed to cache asset", "error", err, "asset_id", assetID)
	}

	s.logger.Info("Asset uploaded successfully", "asset_id", assetID, "asset_url", assetURL)
	return savedAsset, nil
}

// GetAssetByID retrieves an asset by its ID
func (s *AssetsService) GetAssetByID(ctx context.Context, assetID string) (*domain.Asset, error) {
	s.logger.Info("Getting asset by ID", "asset_id", assetID)

	// Check cache first
	asset := new(domain.Asset)
	err := s.cacheService.Get(ctx, assetID, asset)
	if err == nil {
		return asset, nil
	}
	asset, err = s.assetsRepo.GetAssetByID(ctx, assetID)
	if err != nil {
		s.logger.Error("Failed to get asset by ID", "error", err, "asset_id", assetID)
		return nil, domain.NewDomainError(domain.ResourceNotFoundError, "Asset not found", err)
	}

	return asset, nil
}

// GetAssetsByUserID retrieves assets for a specific user
func (s *AssetsService) GetAssetsByUserID(ctx context.Context, userID string, limit, offset int32) ([]*domain.Asset, int32, error) {
	s.logger.Info("Getting assets by user ID", "user_id", userID, "limit", limit, "offset", offset)

	assets, total, err := s.assetsRepo.GetAssetsByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get assets by user ID", "error", err, "user_id", userID)
		return nil, 0, domain.NewDomainError(domain.ResourceNotFoundError, "Failed to get assets", err)
	}

	return assets, total, nil
}

// DeleteAsset deletes an asset by its ID
func (s *AssetsService) DeleteAsset(ctx context.Context, assetID string, userID string) error {
	s.logger.Info("Deleting asset", "asset_id", assetID, "user_id", userID)

	// First, verify the asset belongs to the user
	asset, err := s.assetsRepo.GetAssetByID(ctx, assetID)
	if err != nil {
		s.logger.Error("Failed to get asset for deletion", "error", err, "asset_id", assetID)
		return domain.NewDomainError(domain.ResourceNotFoundError, "Asset not found", err)
	}

	if asset.UserID != nil && *asset.UserID != userID {
		s.logger.Warn("Unauthorized delete attempt", "asset_id", assetID, "user_id", userID, "asset_owner", asset.UserID)
		return domain.NewDomainError(domain.UnauthorizedError, "Asset does not belong to user", nil)
	}

	// Delete from storage (TODO: implement actual storage deletion)
	// storageService.Delete(asset.AssetURL)

	// Delete from database
	if err := s.assetsRepo.DeleteAsset(ctx, assetID); err != nil {
		s.logger.Error("Failed to delete asset", "error", err, "asset_id", assetID)
		return domain.NewDomainError(domain.UnableToDeleteError, "Failed to delete asset", err)
	}

	s.logger.Info("Asset deleted successfully", "asset_id", assetID)
	return nil
}
