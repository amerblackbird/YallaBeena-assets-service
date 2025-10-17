package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"assets-service/internal/core/domain"
	"assets-service/internal/ports"

	config "assets-service/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	Region          string `json:"region"`
	UseSSL          bool   `json:"use_ssl"`
}

// MinIOStorage implements the StorageService interface using MinIO
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	logger     ports.Logger
	config     config.StorageConfig
}

// NewMinIOStorage creates a new MinIO storage service
func NewMinIOStorage(conf config.StorageConfig, logger ports.Logger) (ports.StoragesService, error) {
	// Initialize MinIO client
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: conf.UseSSL,
		Region: conf.Region,
	})
	if err != nil {
		return nil, domain.NewDomainError(domain.BucketConnectionError, "failed to create MinIO client", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: conf.BucketName,
		logger:     logger,
		config:     conf,
	}

	// Ensure bucket exists
	ctx := context.Background()
	err = storage.ensureBucketExists(ctx)
	if err != nil {
		return nil, domain.NewDomainError(domain.ResourceNotFoundError, "failed to ensure bucket exists", err)
	}

	return storage, nil
}

// ensureBucketExists creates the bucket if it doesn't exist
func (s *MinIOStorage) ensureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return domain.NewDomainError(domain.ResourceNotFoundError, "failed to check if bucket exists", err)
	}

	if !exists {
		s.logger.Info("Creating bucket", "bucket", s.bucketName)
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if err != nil {
			return domain.NewDomainError(domain.UnableToCreateError, "failed to create bucket", err)
		}
		s.logger.Info("Bucket created successfully", "bucket", s.bucketName)
	}

	return nil
}

// UploadFile uploads a file to MinIO and returns the URL
func (s *MinIOStorage) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	s.logger.Info("Uploading file to MinIO", "key", key, "size", len(data), "content_type", contentType)

	// Create a reader from the data
	reader := bytes.NewReader(data)

	// Set upload options
	options := minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"uploaded-by": "assets-service",
		},
	}

	// Upload the file
	info, err := s.client.PutObject(ctx, s.bucketName, key, reader, int64(len(data)), options)
	if err != nil {
		s.logger.Error("Failed to upload file to MinIO", "error", err, "key", key)
		return "", domain.NewDomainError(domain.UnableToUploadError, "failed to upload file", err)
	}

	s.logger.Info("File uploaded successfully", "key", key, "etag", info.ETag, "size", info.Size)

	// Generate the file URL
	url := s.generateFileURL(key)
	return url, nil
}

// DeleteFile deletes a file from MinIO
func (s *MinIOStorage) DeleteFile(ctx context.Context, key string) error {
	s.logger.Info("Deleting file from MinIO", "key", key)

	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to delete file from MinIO", "error", err, "key", key)
		return domain.NewDomainError(domain.UnableToDeleteError, "failed to delete file", err)
	}

	s.logger.Info("File deleted successfully", "key", key)
	return nil
}

// GetFileURL returns the URL for accessing a file
func (s *MinIOStorage) GetFileURL(ctx context.Context, key string) (string, error) {
	// For public access, you might want to generate a presigned URL
	// For now, we'll return the direct URL
	url := s.generateFileURL(key)
	return url, nil
}

// // GetFileMetadata returns metadata about a file
// func (s *MinIOStorage) GetFileMetadata(ctx context.Context, key string) (*ports.FileMetadata, error) {
// 	s.logger.Info("Getting file metadata from MinIO", "key", key)

// 	objInfo, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
// 	if err != nil {
// 		s.logger.Error("Failed to get file metadata from MinIO", "error", err, "key", key)
//
// 		return nil, fmt.Errorf("failed to get file metadata: %w", err)
// 	}

// 	metadata := &ports.FileMetadata{
// 		Key:         key,
// 		Size:        objInfo.Size,
// 		ContentType: objInfo.ContentType,
// 		ETag:        objInfo.ETag,
// 		LastMod:     objInfo.LastModified,
// 	}

// 	return metadata, nil
// }

// generateFileURL creates a URL for accessing the file
func (s *MinIOStorage) generateFileURL(key string) string {
	protocol := "http"
	if s.config.UseSSL {
		protocol = "https"
	}

	// Remove any leading slashes from key
	key = strings.TrimPrefix(key, "/")

	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.config.Endpoint, s.bucketName, key)
}

// GeneratePresignedURL generates a presigned URL for temporary access
func (s *MinIOStorage) GeneratePresignedURL(ctx context.Context, key string, expiry int) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, key,
		time.Duration(expiry)*time.Second, nil)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", "error", err, "key", key)
		return "", domain.NewDomainError(domain.UnableToFetchError, "failed to generate presigned URL", err)
	}
	return url.String(), nil
}

func (s *MinIOStorage) Serve(ctx context.Context, w http.ResponseWriter, key string) error {
	object, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to get file from MinIO", "error", err, "key", key)
		return domain.NewDomainError(domain.UnableToFetchError, "failed to get file", err)
	}

	defer object.Close()

	stat, err := object.Stat()
	if err != nil {
		return domain.NewDomainError(domain.UnableToFetchError, "failed to get file stat", err)
	}

	// Set headers for browser download/view
	w.Header().Set("Content-Type", stat.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Stream the file to the client
	if _, err := io.Copy(w, object); err != nil {
		log.Printf("Error writing file to response: %v", err)
	}
	return nil
}
