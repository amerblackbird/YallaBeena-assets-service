package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "assets-service/configs"
	grpcHandler "assets-service/internal/adapters/grpc"
	httpHandler "assets-service/internal/adapters/http"
	kafkaadapter "assets-service/internal/adapters/kafka"
	"assets-service/internal/adapters/logger"
	storageadaper "assets-service/internal/adapters/minio"
	"assets-service/internal/adapters/postgres"
	"assets-service/internal/adapters/redis"
	"assets-service/internal/core/services"

	pb "assets-service/proto/gen/proto"

	"github.com/gorilla/mux"

	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger, err := logger.NewProductionZapLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Initialize database connection
	db, err := postgres.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Close db connection on exit
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Error("Failed to close database connection", "error", err)
		} else {
			appLogger.Info("Database connection closed")
		}
	}()

	// Cache service initialization
	cacheClient := redis.NewRedisClient(cfg.Redis)

	cacheService := redis.NewRedisCacheService(cacheClient, appLogger)

	defer func() {
		if err := cacheService.Close(); err != nil {
			appLogger.Error("Failed to close Redis client", "error", err)
		} else {
			appLogger.Info("Redis client closed")
		}
	}()

	// Initialize repositories
	assetsRepo := postgres.NewAssetsRepository(db, appLogger)

	eventPublisher := kafkaadapter.NewEventPublisher(cfg.Kafka, appLogger)
	eventConsumer := kafkaadapter.NewEventConsumer(cfg.Kafka, appLogger)

	storageService, err := storageadaper.NewMinIOStorage(cfg.Storage, appLogger)
	if err != nil {
		// Stop execution if storage service fails to initialize
		log.Fatalf("Failed to initialize storage service: %v", err)
	}

	assetsService := services.NewAssetsService(assetsRepo, storageService, eventPublisher, cacheService, appLogger)

	// Initialize event handlers
	eventHandlers := kafkaadapter.NewEventHandlers(assetsService, eventPublisher, appLogger)
	eventHandlers.RegisterHandlers(eventConsumer)

	// Initialize HTTP handler
	httpHandlerInstance := httpHandler.NewHTTPHandler(assetsService, storageService, appLogger)

	// Initialize gRPC handler
	grpcServer := grpc.NewServer()
	grpcHandlerInstance := grpcHandler.NewServer(assetsService, appLogger)

	// Setup routes
	r := mux.NewRouter()

	httpHandlerInstance.SetupRoutes(r)

	// Create HTTP server
	httpAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	// Create gRPC server
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC address %s: %v", grpcAddr, err)
	}

	// Register gRPC service
	pb.RegisterAssetsServiceServer(grpcServer, grpcHandlerInstance)

	// Start event consumer in a goroutine
	ctx := context.Background()
	appLogger.Info("Event consumer starting")
	if err := eventConsumer.Start(ctx); err != nil {
		appLogger.Error("Failed to start event consumer", "error", err)
	}

	// Start HTTP server in a goroutine
	go func() {
		appLogger.Info("HTTP Server starting", "address", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server failed to start: %v", err)
		}
	}()

	// Start gRPC server in a goroutine
	go func() {
		appLogger.Info("gRPC Server starting", "address", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server shutting down...")

	// Create a deadline to wait for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop event consumer
	if err := eventConsumer.Stop(); err != nil {
		appLogger.Error("Error stopping event consumer", "error", err)
	}

	if err := eventPublisher.Close(); err != nil {
		appLogger.Error("Error closing event publisher", "error", err)
	} else {
		appLogger.Info("Event publisher closed")
	}

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP Server forced to shutdown: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	appLogger.Info("Servers exited")

}
