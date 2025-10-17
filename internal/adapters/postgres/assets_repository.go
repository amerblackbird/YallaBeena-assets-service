package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"assets-service/internal/core/domain"
	"assets-service/internal/ports"
)

// AssetsRepository implements the assets repository interface for PostgreSQL
type AssetsRepository struct {
	db     *sql.DB
	logger ports.Logger
}

// NewAssetsRepository creates a new assets repository
func NewAssetsRepository(db *sql.DB, logger ports.Logger) ports.AssetsRepository {
	return &AssetsRepository{
		db:     db,
		logger: logger,
	}
}

// CreateAsset creates a new asset in the database
func (r *AssetsRepository) CreateAsset(ctx context.Context, asset *domain.CreateAssetDto) (*domain.Asset, error) {
	query := `
		INSERT INTO assets (url, filename, file_size, metadata, secure, storage_key, 
			storage_provider, resource_id, resource_type, content_type, user_id, access_level, 
			allowed_roles, is_encrypted, encryption_key, tags, file_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, url, public_url, filename, file_size, metadata, secure, storage_key, 
			storage_provider, resource_id, resource_type, content_type, user_id, access_level, 
			allowed_roles, is_encrypted, encryption_key, last_accessed_at, deleted_at, tags, 
			created_at, updated_at, active, file_hash
	`

	row := r.db.QueryRowContext(ctx, query,
		asset.URL,
		asset.Filename,
		asset.FileSize,
		asset.Metadata,
		asset.Secure,
		asset.StorageKey,
		asset.StorageProvider,
		asset.ResourceID,
		asset.ResourceType,
		asset.ContentType,
		asset.UserID,
		asset.AccessLevel,
		asset.AllowedRoles,
		asset.IsEncrypted,
		asset.EncryptionKey,
		asset.Tags,
		asset.FileHash,
	)

	var createdAsset domain.Asset
	err := row.Scan(
		&createdAsset.ID,
		&createdAsset.URL,
		&createdAsset.PublicURL,
		&createdAsset.Filename,
		&createdAsset.FileSize,
		&createdAsset.Metadata,
		&createdAsset.Secure,
		&createdAsset.StorageKey,
		&createdAsset.StorageProvider,
		&createdAsset.ResourceID,
		&createdAsset.ResourceType,
		&createdAsset.ContentType,
		&createdAsset.UserID,
		&createdAsset.AccessLevel,
		&createdAsset.AllowedRoles,
		&createdAsset.IsEncrypted,
		&createdAsset.EncryptionKey,
		&createdAsset.LastAccessedAt,
		&createdAsset.DeletedAt,
		&createdAsset.Tags,
		&createdAsset.CreatedAt,
		&createdAsset.UpdatedAt,
		&createdAsset.Active,
		&createdAsset.FileHash,
	)

	if err != nil {
		r.logger.Error("Failed to create asset", "error", err)
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return &createdAsset, nil
}

// GetAssetByID retrieves an asset by its ID
func (r *AssetsRepository) GetAssetByID(ctx context.Context, assetID string) (*domain.Asset, error) {
	query := `
		SELECT id, url, public_url, filename, file_size, metadata, secure, storage_key, 
			storage_provider, resource_id, resource_type, content_type, user_id, access_level, 
			allowed_roles, is_encrypted, encryption_key, last_accessed_at, deleted_at, tags, 
			created_at, updated_at, active, file_hash
		FROM assets
		WHERE id = $1 AND active = true AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, assetID)

	var asset domain.Asset
	err := row.Scan(
		&asset.ID,
		&asset.URL,
		&asset.PublicURL,
		&asset.Filename,
		&asset.FileSize,
		&asset.Metadata,
		&asset.Secure,
		&asset.StorageKey,
		&asset.StorageProvider,
		&asset.ResourceID,
		&asset.ResourceType,
		&asset.ContentType,
		&asset.UserID,
		&asset.AccessLevel,
		&asset.AllowedRoles,
		&asset.IsEncrypted,
		&asset.EncryptionKey,
		&asset.LastAccessedAt,
		&asset.DeletedAt,
		&asset.Tags,
		&asset.CreatedAt,
		&asset.UpdatedAt,
		&asset.Active,
		&asset.FileHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		r.logger.Error("Failed to get asset by ID", "error", err, "asset_id", assetID)
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return &asset, nil
}

// GetAssetsByUserID retrieves assets for a specific user with pagination
func (r *AssetsRepository) GetAssetsByUserID(ctx context.Context, userID string, limit, offset int32) ([]*domain.Asset, int32, error) {
	// First, get the total count
	countQuery := `
		SELECT COUNT(*)
		FROM assets
		WHERE user_id = $1 AND active = true AND deleted_at IS NULL
	`

	var totalCount int32
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to count assets", "error", err, "user_id", userID)
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	// Then get the actual assets
	query := `
		SELECT id, url, public_url, filename, file_size, metadata, secure, storage_key, 
			storage_provider, resource_id, resource_type, content_type, user_id, access_level, 
			allowed_roles, is_encrypted, encryption_key, last_accessed_at, deleted_at, tags, 
			created_at, updated_at, active, file_hash
		FROM assets
		WHERE user_id = $1 AND active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("Failed to get assets by user ID", "error", err, "user_id", userID)
		return nil, 0, fmt.Errorf("failed to get assets: %w", err)
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var asset domain.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.URL,
			&asset.PublicURL,
			&asset.Filename,
			&asset.FileSize,
			&asset.Metadata,
			&asset.Secure,
			&asset.StorageKey,
			&asset.StorageProvider,
			&asset.ResourceID,
			&asset.ResourceType,
			&asset.ContentType,
			&asset.UserID,
			&asset.AccessLevel,
			&asset.AllowedRoles,
			&asset.IsEncrypted,
			&asset.EncryptionKey,
			&asset.LastAccessedAt,
			&asset.DeletedAt,
			&asset.Tags,
			&asset.CreatedAt,
			&asset.UpdatedAt,
			&asset.Active,
			&asset.FileHash,
		)
		if err != nil {
			r.logger.Error("Failed to scan asset", "error", err)
			return nil, 0, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, &asset)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return assets, totalCount, nil
}

// UpdateAsset updates an existing asset
func (r *AssetsRepository) UpdateAsset(ctx context.Context, asset *domain.UpdateAssetDto) (*domain.Asset, error) {
	// Build dynamic query based on provided fields
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 2

	if asset.URL != nil {
		setParts = append(setParts, fmt.Sprintf("url = $%d", argIndex))
		args = append(args, *asset.URL)
		argIndex++
	}
	if asset.PublicURL != nil {
		setParts = append(setParts, fmt.Sprintf("public_url = $%d", argIndex))
		args = append(args, *asset.PublicURL)
		argIndex++
	}
	if asset.Filename != nil {
		setParts = append(setParts, fmt.Sprintf("filename = $%d", argIndex))
		args = append(args, *asset.Filename)
		argIndex++
	}
	if asset.FileSize != nil {
		setParts = append(setParts, fmt.Sprintf("file_size = $%d", argIndex))
		args = append(args, *asset.FileSize)
		argIndex++
	}
	if asset.Metadata != nil {
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, asset.Metadata)
		argIndex++
	}
	if asset.Secure != nil {
		setParts = append(setParts, fmt.Sprintf("secure = $%d", argIndex))
		args = append(args, *asset.Secure)
		argIndex++
	}
	if asset.StorageKey != nil {
		setParts = append(setParts, fmt.Sprintf("storage_key = $%d", argIndex))
		args = append(args, *asset.StorageKey)
		argIndex++
	}
	if asset.StorageProvider != nil {
		setParts = append(setParts, fmt.Sprintf("storage_provider = $%d", argIndex))
		args = append(args, *asset.StorageProvider)
		argIndex++
	}
	if asset.ResourceID != nil {
		setParts = append(setParts, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, *asset.ResourceID)
		argIndex++
	}
	if asset.ResourceType != nil {
		setParts = append(setParts, fmt.Sprintf("resource_type = $%d", argIndex))
		args = append(args, *asset.ResourceType)
		argIndex++
	}
	if asset.ContentType != nil {
		setParts = append(setParts, fmt.Sprintf("content_type = $%d", argIndex))
		args = append(args, *asset.ContentType)
		argIndex++
	}
	if asset.UserID != nil {
		setParts = append(setParts, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *asset.UserID)
		argIndex++
	}
	if asset.AccessLevel != nil {
		setParts = append(setParts, fmt.Sprintf("access_level = $%d", argIndex))
		args = append(args, *asset.AccessLevel)
		argIndex++
	}
	if asset.AllowedRoles != nil {
		setParts = append(setParts, fmt.Sprintf("allowed_roles = $%d", argIndex))
		args = append(args, asset.AllowedRoles)
		argIndex++
	}
	if asset.IsEncrypted != nil {
		setParts = append(setParts, fmt.Sprintf("is_encrypted = $%d", argIndex))
		args = append(args, *asset.IsEncrypted)
		argIndex++
	}
	if asset.EncryptionKey != nil {
		setParts = append(setParts, fmt.Sprintf("encryption_key = $%d", argIndex))
		args = append(args, *asset.EncryptionKey)
		argIndex++
	}
	if asset.Tags != nil {
		setParts = append(setParts, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, asset.Tags)
		argIndex++
	}
	if asset.FileHash != "" {
		setParts = append(setParts, fmt.Sprintf("file_hash = $%d", argIndex))
		args = append(args, asset.FileHash)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE assets
		SET %s
		WHERE id = $1 AND active = true AND deleted_at IS NULL
		RETURNING id, url, public_url, filename, file_size, metadata, secure, storage_key, 
			storage_provider, resource_id, resource_type, content_type, user_id, access_level, 
			allowed_roles, is_encrypted, encryption_key, last_accessed_at, deleted_at, tags, 
			created_at, updated_at, active, file_hash
	`, strings.Join(setParts, ", "))

	// Prepend the ID parameter
	finalArgs := append([]interface{}{asset.ID}, args...)

	row := r.db.QueryRowContext(ctx, query, finalArgs...)

	var updatedAsset domain.Asset
	err := row.Scan(
		&updatedAsset.ID,
		&updatedAsset.URL,
		&updatedAsset.PublicURL,
		&updatedAsset.Filename,
		&updatedAsset.FileSize,
		&updatedAsset.Metadata,
		&updatedAsset.Secure,
		&updatedAsset.StorageKey,
		&updatedAsset.StorageProvider,
		&updatedAsset.ResourceID,
		&updatedAsset.ResourceType,
		&updatedAsset.ContentType,
		&updatedAsset.UserID,
		&updatedAsset.AccessLevel,
		&updatedAsset.AllowedRoles,
		&updatedAsset.IsEncrypted,
		&updatedAsset.EncryptionKey,
		&updatedAsset.LastAccessedAt,
		&updatedAsset.DeletedAt,
		&updatedAsset.Tags,
		&updatedAsset.CreatedAt,
		&updatedAsset.UpdatedAt,
		&updatedAsset.Active,
		&updatedAsset.FileHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset not found")
		}
		r.logger.Error("Failed to update asset", "error", err, "asset_id", asset.ID)
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	return &updatedAsset, nil
}

// DeleteAsset soft deletes an asset by setting deleted_at timestamp
func (r *AssetsRepository) DeleteAsset(ctx context.Context, assetID string) error {
	query := `
		UPDATE assets
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND active = true AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, assetID)
	if err != nil {
		r.logger.Error("Failed to delete asset", "error", err, "asset_id", assetID)
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}

// GetAssetsByFilter retrieves assets based on filters with pagination
func (r *AssetsRepository) GetAssetsByFilter(ctx context.Context, filter *domain.AssetFilter) ([]*domain.Asset, int32, error) {
	whereClauses := []string{"active = true", "deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause dynamically
	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}
	if filter.ContentType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("content_type = $%d", argIndex))
		args = append(args, *filter.ContentType)
		argIndex++
	}
	if filter.ResourceType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("resource_type = $%d", argIndex))
		args = append(args, *filter.ResourceType)
		argIndex++
	}
	if filter.ResourceID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, *filter.ResourceID)
		argIndex++
	}
	if filter.AccessLevel != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("access_level = $%d", argIndex))
		args = append(args, *filter.AccessLevel)
		argIndex++
	}
	if filter.Secure != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("secure = $%d", argIndex))
		args = append(args, *filter.Secure)
		argIndex++
	}
	if filter.IsEncrypted != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_encrypted = $%d", argIndex))
		args = append(args, *filter.IsEncrypted)
		argIndex++
	}
	if filter.StorageProvider != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("storage_provider = $%d", argIndex))
		args = append(args, *filter.StorageProvider)
		argIndex++
	}
	if len(filter.Tags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("tags && $%d", argIndex))
		args = append(args, filter.Tags)
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM assets WHERE %s", whereClause)
	var totalCount int32
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to count assets with filter", "error", err)
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	// Main query with limit and offset
	limitArgs := append(args, filter.Limit, filter.Offset)
	query := fmt.Sprintf(`
		SELECT id, url, public_url, filename, file_size, metadata, secure, storage_key, 
			storage_provider, resource_id, resource_type, content_type, user_id, access_level, 
			allowed_roles, is_encrypted, encryption_key, last_accessed_at, deleted_at, tags, 
			created_at, updated_at, active, file_hash
		FROM assets 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := r.db.QueryContext(ctx, query, limitArgs...)
	if err != nil {
		r.logger.Error("Failed to get assets with filter", "error", err)
		return nil, 0, fmt.Errorf("failed to get assets: %w", err)
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var asset domain.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.URL,
			&asset.PublicURL,
			&asset.Filename,
			&asset.FileSize,
			&asset.Metadata,
			&asset.Secure,
			&asset.StorageKey,
			&asset.StorageProvider,
			&asset.ResourceID,
			&asset.ResourceType,
			&asset.ContentType,
			&asset.UserID,
			&asset.AccessLevel,
			&asset.AllowedRoles,
			&asset.IsEncrypted,
			&asset.EncryptionKey,
			&asset.LastAccessedAt,
			&asset.DeletedAt,
			&asset.Tags,
			&asset.CreatedAt,
			&asset.UpdatedAt,
			&asset.Active,
			&asset.FileHash,
		)
		if err != nil {
			r.logger.Error("Failed to scan asset", "error", err)
			return nil, 0, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, &asset)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Row iteration error", "error", err)
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return assets, totalCount, nil
}

// UpdateLastAccessedAt updates the last_accessed_at timestamp for an asset
func (r *AssetsRepository) UpdateLastAccessedAt(ctx context.Context, assetID string) error {
	query := `
		UPDATE assets 
		SET last_accessed_at = NOW(), updated_at = NOW() 
		WHERE id = $1 AND active = true AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, assetID)
	if err != nil {
		r.logger.Error("Failed to update last accessed time", "error", err, "asset_id", assetID)
		return fmt.Errorf("failed to update last accessed time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}
