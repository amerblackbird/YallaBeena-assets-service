package grpc

import (
	"context"
	"encoding/json"
	"time"

	"assets-service/internal/core/domain"
	"assets-service/internal/ports"
	pb "assets-service/proto/gen/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the gRPC server for assets service
type Server struct {
	pb.UnimplementedAssetsServiceServer
	assetsService ports.AssetsService
	logger        ports.Logger
}

// NewServer creates a new gRPC server
func NewServer(assetsService ports.AssetsService, logger ports.Logger) *Server {
	return &Server{
		assetsService: assetsService,
		logger:        logger,
	}
}

// HealthCheck returns the service health status
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	s.logger.Info("gRPC HealthCheck called")

	return &pb.HealthCheckResponse{
		Status:  "healthy",
		Service: "assets-service",
		Version: "1.0.0",
	}, nil
}

// UploadAsset uploads a new asset and returns metadata
func (s *Server) UploadAsset(ctx context.Context, req *pb.UploadAssetRequest) (*pb.UploadAssetResponse, error) {
	s.logger.Info("gRPC UploadAsset called", "filename", req.Filename, "user_id", req.UserId)

	// Validate request
	if req.Filename == "" || req.UserId == "" || len(req.FileData) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "filename, user_id, and file_data are required")
	}

	var resourceId *string
	if req.ResourceId != "" {
		resourceId = &req.ResourceId
	}

	var resourceType *string
	if req.ResouceType != "" {
		resourceType = &req.ResouceType
	}

	meta := req.Metadata
	var jsonMeta json.RawMessage
	if meta != nil {
		bytes, err := json.Marshal(meta)
		if err != nil {
			s.logger.Error("Failed to marshal metadata", "error", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid metadata format: %v", err)
		}
		jsonMeta = bytes
	}

	s.logger.Info("Resource2", "ResourceType2", req.ResouceType, "ResourceID", req.ResourceId)

	// Convert gRPC request to domain DTO
	createDto := &domain.CreateAssetDto{
		Filename:        req.Filename,
		ContentType:     req.ContentType,
		FileSize:        int64(len(req.FileData)),
		UserID:          &req.UserId,
		Metadata:        jsonMeta,
		Secure:          false,
		Tags:            []string{},
		AccessLevel:     "private",
		AllowedRoles:    []string{},
		IsEncrypted:     false,
		EncryptionKey:   nil,
		StorageProvider: nil,
		ResourceID:      resourceId,
		ResourceType:    resourceType,
	}

	// Call the service
	asset, err := s.assetsService.UploadAsset(ctx, createDto, req.FileData)
	if err != nil {
		s.logger.Error("Failed to upload asset", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to upload asset: %v", err)
	}

	// Convert domain model to gRPC response
	pbAsset := s.assetDomainToProto(asset)

	return &pb.UploadAssetResponse{
		Asset: pbAsset,
	}, nil
}

// GetAsset retrieves an asset by its ID
func (s *Server) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.GetAssetResponse, error) {
	s.logger.Info("gRPC GetAsset called", "asset_id", req.AssetId)

	asset, err := s.assetsService.GetAssetByID(ctx, req.AssetId)
	if err != nil {
		s.logger.Error("Failed to get asset by ID", "error", err, "asset_id", req.AssetId)
		return nil, status.Errorf(codes.NotFound, "asset not found: %v", err)
	}

	pbAsset := s.assetDomainToProto(asset)

	return &pb.GetAssetResponse{
		Asset: pbAsset,
	}, nil
}

// GetAssetsByUser retrieves assets for a specific user
func (s *Server) GetAssetsByUser(ctx context.Context, req *pb.GetAssetsByUserRequest) (*pb.GetAssetsByUserResponse, error) {
	s.logger.Info("gRPC GetAssetsByUser called", "user_id", req.UserId)

	assets, total, err := s.assetsService.GetAssetsByUserID(ctx, req.UserId, req.Limit, req.Offset)
	if err != nil {
		s.logger.Error("Failed to get assets by user ID", "error", err, "user_id", req.UserId)
		return nil, status.Errorf(codes.Internal, "failed to get assets: %v", err)
	}

	pbAssets := make([]*pb.Asset, len(assets))
	for i, asset := range assets {
		pbAssets[i] = s.assetDomainToProto(asset)
	}

	return &pb.GetAssetsByUserResponse{
		Assets:     pbAssets,
		TotalCount: total,
	}, nil
}

// DeleteAsset deletes an asset by its ID
func (s *Server) DeleteAsset(ctx context.Context, req *pb.DeleteAssetRequest) (*pb.DeleteAssetResponse, error) {
	s.logger.Info("gRPC DeleteAsset called", "asset_id", req.AssetId, "user_id", req.UserId)

	err := s.assetsService.DeleteAsset(ctx, req.AssetId, req.UserId)
	if err != nil {
		s.logger.Error("Failed to delete asset", "error", err, "asset_id", req.AssetId)
		return nil, status.Errorf(codes.Internal, "failed to delete asset: %v", err)
	}

	return &pb.DeleteAssetResponse{
		Success: true,
		Message: "Asset deleted successfully",
	}, nil
}

// assetDomainToProto converts a domain Asset to protobuf Asset
func (s *Server) assetDomainToProto(asset *domain.Asset) *pb.Asset {
	userId := ""
	if asset.UserID != nil {
		userId = *asset.UserID
	}
	resourceId := ""
	if asset.ResourceID != nil {
		resourceId = *asset.ResourceID
	}
	resourceType := ""
	if asset.ResourceType != nil {
		resourceType = *asset.ResourceType
	}
	pbAsset := &pb.Asset{
		AssetId:     asset.ID.String(),
		AssetUrl:    asset.URL,
		PublicUrl:   asset.PublicURL,
		Filename:    asset.Filename,
		ContentType: asset.ContentType,
		FileSize:    asset.FileSize,
		UserId:      userId,
		ResouceType: resourceType,
		ResourceId:  resourceId,
		Secure:      asset.Secure,
		AccessLevel: asset.AccessLevel,
	}

	// Convert string timestamps to timestamppb.Timestamp
	if asset.CreatedAt != "" {
		if createdAt, err := time.Parse(time.RFC3339, asset.CreatedAt); err == nil {
			pbAsset.CreatedAt = timestamppb.New(createdAt)
		}
	}
	if asset.UpdatedAt != "" {
		if updatedAt, err := time.Parse(time.RFC3339, asset.UpdatedAt); err == nil {
			pbAsset.UpdatedAt = timestamppb.New(updatedAt)
		}
	}

	return pbAsset
}
