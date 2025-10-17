# Assets Service

A microservice for managing activity logs with both HTTP and gRPC APIs.

## Features

- **HTTP API**: RESTful endpoints for activity log management
- **gRPC API**: High-performance RPC interface
- **Event-driven architecture**: Kafka integration for event publishing/consuming
- **Caching**: Redis for performance optimization
- **Database**: PostgreSQL for persistent storage

## APIs

### HTTP API

- **Port**: 8080 (configurable via `SERVER_PORT`)
- **Base URL**: `http://localhost:8080`

### gRPC API

- **Port**: 9090 (configurable via `GRPC_PORT`)
- **Address**: `localhost:9090`

#### gRPC Service Methods:

- `LogActivity(CreateActivityLogRequest) returns (CreateActivityLogResponse)`
- `GetActivityLogByID(GetActivityLogByIDRequest) returns (GetActivityLogByIDResponse)`
- `GetActivityLogsByUserID(GetActivityLogsByUserIDRequest) returns (GetActivityLogsByUserIDResponse)`

## Configuration

The service can be configured using environment variables:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
GRPC_PORT=9090

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=assets_service
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=assets_service
KAFKA_TOPIC_ACTIVITY_LOG_EVENTS=activity.logs
```

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL
- Redis
- Kafka
- Protocol Buffers compiler (`protoc`)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   make install-proto  # Install protobuf dependencies (macOS)
   make tidy          # Install Go dependencies
   ```

### Building

```bash
# Build the main application
make build

# Generate protobuf code
make proto

# Build gRPC client example
make client
```

### Running

```bash
# Run the service
make run

# Or run directly
./assets-service
```

### Testing gRPC

Use the provided client example:

```bash
# Build and run the gRPC client
make client
./grpc_client
```

### Docker

Build and run with Docker:

```bash
docker build -t assets-service .
docker run -p 8080:8080 -p 9090:9090 assets-service
```

## Project Structure

```
.
├── cmd/app/                 # Application entry point
├── configs/                 # Configuration management
├── internal/
│   ├── adapters/           # External adapters
│   │   ├── grpc/          # gRPC server implementation
│   │   ├── http/          # HTTP handlers
│   │   ├── kafka/         # Kafka event handling
│   │   ├── postgres/      # Database repositories
│   │   └── redis/         # Cache implementation
│   ├── core/
│   │   ├── domain/        # Domain models and DTOs
│   │   ├── events/        # Domain events
│   │   └── services/      # Business logic
│   └── ports/             # Interfaces/contracts
├── proto/                  # Protocol buffer definitions
│   └── gen/               # Generated protobuf code
├── examples/              # Example clients
└── migrations/            # Database migrations
```

## API Examples

### gRPC Client (Go)

```go
import (
    pb "assets-service/proto/gen/proto"
    "google.golang.org/grpc"
)

conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := pb.NewActivityLogsServiceClient(conn)

resp, err := client.LogActivity(ctx, &pb.CreateActivityLogRequest{
    UserId:   "user123",
    Action:   "login",
    Resource: "auth",
    Details:  "User logged in successfully",
})
```

### Using grpcurl

```bash
# List services
grpcurl -plaintext localhost:9090 list

# Call LogActivity
grpcurl -plaintext -d '{
  "user_id": "user123",
  "action": "login",
  "resource": "auth",
  "details": "User logged in"
}' localhost:9090 activity_logs.ActivityLogsService/LogActivity
```
