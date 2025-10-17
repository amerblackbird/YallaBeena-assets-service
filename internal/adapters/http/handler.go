package http

import (
	"context"
	"net/http"
	"strings"

	domain "assets-service/internal/core/domain"
	ports "assets-service/internal/ports"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPHandler implements the HTTP adapter for the activity logs service
type HTTPHandler struct {
	assetsService  ports.AssetsService
	storageService ports.StoragesService
	logger         ports.Logger
	Validator      validator.Validate
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(
	assetsService ports.AssetsService,
	storageService ports.StoragesService,
	logger ports.Logger) ports.HTTPHandler {
	return &HTTPHandler{
		assetsService:  assetsService,
		storageService: storageService,
		logger:         logger,
		Validator:      *domain.NewValidator(),
	}
}

func (h *HTTPHandler) SetupRoutes(r *mux.Router) {
	// Health check endpoint
	r.HandleFunc("/health", h.handleHealth).Methods("GET")

	// Define your HTTP routes here
	r.HandleFunc("/assets/{id}", h.handleGetAssetById).Methods("GET")

	// Log all routes
	if err := h.ShowRoutes(r); err != nil {
		h.logger.Error("Failed to show routes", zap.Error(err))
	}
}

func (h *HTTPHandler) ShowRoutes(r *mux.Router) error {

	var routes []string
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			// If no methods are defined, skip this route or use a default
			routes = append(routes, "ALL "+pathTemplate)
			return nil
		}

		routes = append(routes, strings.Join(methods, ",")+" "+pathTemplate)
		return nil
	})
	if err != nil {
		return err
	}

	for _, route := range routes {
		parts := strings.SplitN(route, " ", 2)
		if len(parts) == 2 {
			h.logger.Info("Registered HTTP route",
				zap.String("methods", parts[0]),
				zap.String("path", parts[1]),
				zap.String("service", "assets-service"))
		} else {
			h.logger.Info("Registered HTTP route",
				zap.String("route", route),
				zap.String("service", "assets-service"))
		}
	}
	return nil
}

func (h *HTTPHandler) handleGetAssetById(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		h.responseWithError(w, http.StatusBadRequest, domain.NewDomainError(
			domain.ResourceNotFoundError,
			"Missing asset ID", nil))
		return
	}

	asset, err := h.assetsService.GetAssetByID(r.Context(), id)
	if err != nil {
		h.responseWithError(w, http.StatusBadRequest, err)
		return
	}
	if asset == nil {
		h.responseWithError(w, http.StatusBadRequest, domain.NewDomainError(
			domain.ResourceNotFoundError,
			"Asset not found", nil))
		return
	}

	// Serve asset
	// serverUrl := ""
	// assetUrl := fmt.Sprintf("%s/%s", serverUrl, asset.PublicURL)

	if asset.StorageKey == nil || *asset.StorageKey == "" {
		h.responseWithError(w, http.StatusInternalServerError, domain.NewDomainError(
			domain.UnableToFetchError,
			"Asset storage key is missing", nil))
		return
	}

	err = h.storageService.Serve(r.Context(), w, *asset.StorageKey)
	if err != nil {
		h.responseWithError(w, http.StatusInternalServerError, err)
		return
	}

	// // Server
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(fmt.Sprintf(`{"asset_id":"%s","asset_url":"%s","filename":"%s","content_type":"%s","file_size":%d}`,
	// 	asset.ID,
	// 	asset.PublicURL,
	// 	asset.Filename,
	// 	asset.ContentType,
	// 	asset.FileSize,
	// )))

}

func (h *HTTPHandler) HandleGetAssetsByID(ctx context.Context, assetID string) (*domain.Asset, error) {
	return nil, nil
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"assets-service","version":"1.0.0"}`))
}
