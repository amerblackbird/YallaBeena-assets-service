# Build stage
FROM golang:1.23.1-alpine AS builder

# Set working directory
WORKDIR /app

# Install protobuf compiler and dependencies
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    protobuf \
    protobuf-dev && \
    update-ca-certificates

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf code
RUN protoc --go_out=./proto/gen --go_opt=paths=source_relative \
    --go-grpc_out=./proto/gen --go-grpc_opt=paths=source_relative \
    proto/*.proto

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/app/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/


# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migrations directory into the final image
# COPY --from=builder /app/migrations /migrations

# Expose HTTP and gRPC ports
EXPOSE 8090 9090

# Run the application
CMD ["./main"]
