# Makefile for Assets Service

# Variables
APP_NAME = assets-service
PROTO_DIR = proto
PROTO_GEN_DIR = proto/gen

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	go build -o $(APP_NAME) ./cmd/app

# Run the application
.PHONY: run
run: build
	./$(APP_NAME)

# Generate protobuf code
.PHONY: proto
proto:
	protoc --go_out=./$(PROTO_GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=./$(PROTO_GEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

# Build gRPC client example
.PHONY: client
client:
	go build -o grpc_assets_client ./examples/grpc_assets_client.go

# Build old activity client
.PHONY: activity-client
activity-client:
	go build -o grpc_activity_client ./examples/grpc_activity_client.go

# Run tests
.PHONY: test
test:
	go test ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(APP_NAME) grpc_assets_client grpc_activity_client

# Run go mod tidy
.PHONY: tidy
tidy:
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Install protobuf dependencies (macOS)
.PHONY: install-proto
install-proto:
	brew install protobuf
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Test grpc
.PHONY: test-grpc
test-grpc:
	grpcurl -plaintext localhost:9090 list
	grpcurl -plaintext localhost:9090 assets.AssetsService/HealthCheck

# Docker targets
.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 -p 9090:9090 $(APP_NAME):latest

.PHONY: docker-compose-up
docker-compose-up:
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	docker-compose down

.PHONY: docker-test
docker-test:
	./test-docker-grpc.sh
.PHONY: gen-proto
gen-proto:
	protoc --go_out=./proto/gen --go_opt=paths=source_relative --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative proto/assets.proto


# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  proto          - Generate protobuf code"
	@echo "  client         - Build gRPC client example"
	@echo "  test           - Run tests"
	@echo "  test-grpc      - Test gRPC endpoints"
	@echo "  clean          - Clean build artifacts"
	@echo "  tidy           - Run go mod tidy"
	@echo "  fmt            - Format code"
	@echo "  install-proto  - Install protobuf dependencies (macOS)"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-test    - Test gRPC in Docker"
	@echo "  docker-compose-up   - Start all services with docker-compose"
	@echo "  docker-compose-down - Stop all services"
	@echo "  help           - Show this help message"