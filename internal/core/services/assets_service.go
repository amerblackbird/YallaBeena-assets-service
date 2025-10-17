package services

import (
	"context"
	"crypto/sha256"
	"fmt"

	"assets-service/internal/core/domain"
	"assets-service/internal/ports"
	utils "assets-service/internal/utils"
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

	// Add metadata including file hash for integrity
	fileHash := fmt.Sprintf("%x", sha256.Sum256(fileData))
	fileSize := int64(len(fileData))

	// Generate file key for storage (handle null UserID)
	fileKey := createDto.GetStoreKey()
	metadataJSON := createDto.GetMetadata(fileKey, fileHash)

	// Log upload start
	s.logger.Info("Uploading asset", "filename", createDto.Filename, "user_id", createDto.UserID, "file_key", fileKey)

	// Upload file to storage
	assetURL, err := s.storageService.UploadFile(ctx, fileKey, fileData, createDto.ContentType)
	if err != nil {
		s.logger.Error("Failed to upload file to storage", "error", err, "file_key", fileKey)

		return nil, domain.NewDomainError(domain.UnableToMarshalError, "failed to upload file to storag", err)
	}

	s.logger.Info("File uploaded to storage", "file_key", fileKey, "asset_url", assetURL)

	// Create asset DTO for repository
	assetDto := &domain.CreateAssetDto{
		StorageKey:      &fileKey,
		StorageProvider: utils.StringPtr("minio"),
		URL:             assetURL,
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
	asset, err := s.assetsRepo.CreateAsset(ctx, assetDto)
	if err != nil {
		s.logger.Error("Failed to save asset metadata", "error", err)

		// Rollback: delete the file from storage if database save fails
		if deleteErr := s.storageService.DeleteFile(ctx, fileKey); deleteErr != nil {
			s.logger.Error("Failed to rollback file upload", "error", deleteErr, "file_key", fileKey)
		}
		return nil, domain.NewDomainError(domain.UnableToMarshalError, "Failed to save asset metadata", err)
	}

	// Cache the asset
	cacheKey := fmt.Sprintf("assets:%s", asset.ID.String())
	if err := s.cacheService.Set(ctx, cacheKey, asset, 0); err != nil {
		s.logger.Error("Failed to cache asset", "error", err, "domain", "cache")
	}
	s.logger.Info("Asset uploaded successfully", "asset_url", assetURL)
	return asset, nil
}

// GetAssetByID retrieves an asset by its ID
func (s *AssetsService) GetAssetByID(ctx context.Context, assetID string) (*domain.Asset, error) {
	s.logger.Info("Getting asset by ID", "asset_id", assetID)

	// Check cache first
	asset := new(domain.Asset)
	cacheKey := fmt.Sprintf("assets:%s", assetID)
	err := s.cacheService.Get(ctx, cacheKey, asset)
	if err == nil {
		return asset, nil
	}
	asset, err = s.assetsRepo.GetAssetByID(ctx, assetID)
	if err != nil {
		s.logger.Error("Failed to get asset by ID", "error", err, "asset_id", assetID)
		return nil, domain.NewDomainError(domain.ResourceNotFoundError, "Asset not found", err)
	}

	// Cache the asset
	if err := s.cacheService.Set(ctx, cacheKey, asset, 0); err != nil {
		s.logger.Error("Failed to cache asset", "error", err, "domain", "cache")
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

	// Delete from storage
	err = s.storageService.DeleteFile(ctx, *asset.StorageKey)
	if err != nil {
		s.logger.Error("Failed to delete file from storage", "error", err, "storage_key", *asset.StorageKey)
		return domain.NewDomainError(domain.UnableToDeleteError, "Failed to delete file from storage", err)
	}

	// Delete from cache
	cacheKey := fmt.Sprintf("assets:%s", assetID)
	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to delete asset from cache", "error", err, "asset_id", assetID)
	}

	// Delete from database
	if err := s.assetsRepo.DeleteAsset(ctx, assetID); err != nil {
		s.logger.Error("Failed to delete asset", "error", err, "asset_id", assetID)
		return domain.NewDomainError(domain.UnableToDeleteError, "Failed to delete asset", err)
	}

	s.logger.Info("Asset deleted successfully", "asset_id", assetID)
	return nil
}
