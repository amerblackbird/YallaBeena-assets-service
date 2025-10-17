package ports

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"assets-service/internal/core/domain"
	pb "assets-service/proto/gen/proto"
)

// AssetsService defines the interface for asset management
type AssetsService interface {
	UploadAsset(ctx context.Context, createDto *domain.CreateAssetDto, fileData []byte) (*domain.Asset, error)
	GetAssetByID(ctx context.Context, assetID string) (*domain.Asset, error)
	GetAssetsByUserID(ctx context.Context, userID string, limit, offset int32) ([]*domain.Asset, int32, error)
	DeleteAsset(ctx context.Context, assetID string, userID string) error
}

type StoragesService interface {
	UploadFile(ctx context.Context, path string, fileData []byte, contentType string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	Serve(ctx context.Context, w http.ResponseWriter, key string) error
}

type HTTPHandler interface {
	// SetupRoutes sets up the HTTP routes for the handler
	SetupRoutes(router *mux.Router)

	// Show all registered routes
	ShowRoutes(r *mux.Router) error

	// HandleGetAssetsByID retrieves an asset by its ID
	HandleGetAssetsByID(ctx context.Context, assetID string) (*domain.Asset, error)
}

// ActivityLogsService interface (keeping for backwards compatibility)
type RPCHandler interface {
	HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error)
	UploadAsset(ctx context.Context, req *pb.UploadAssetRequest) (*pb.UploadAssetResponse, error)
	GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.GetAssetResponse, error)
	GetAssetsByUser(ctx context.Context, req *pb.GetAssetsByUserRequest) (*pb.GetAssetsByUserResponse, error)
	DeleteAsset(ctx context.Context, req *pb.DeleteAssetRequest) (*pb.DeleteAssetResponse, error)
}
