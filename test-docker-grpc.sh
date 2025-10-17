#!/bin/bash

echo "🐳 Testing gRPC in Docker..."

# Stop any running container
echo "Stopping any existing containers..."
docker stop $(docker ps -q --filter ancestor=assets-service) 2>/dev/null || true

# Run the container with environment variables that point to existing services
echo "Starting container with database connection..."
docker run --rm -d \
  --name assets-service-test \
  --network host \
  -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=postgres \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  -e KAFKA_BROKERS=localhost:9092 \
  -p 8080:8080 \
  -p 9090:9090 \
  assets-service:latest

echo "Waiting for service to start..."
sleep 5

# Test gRPC health check
echo "Testing gRPC HealthCheck..."
if command -v grpcurl &> /dev/null; then
    grpcurl -plaintext localhost:9090 assets.AssetsService/HealthCheck
else
    echo "grpcurl not found. Install with: brew install grpcurl"
    echo "Testing HTTP health check instead..."
    curl -s http://localhost:8090/health | jq . || curl -s http://localhost:8090/health
fi

echo "Stopping test container..."
docker stop assets-service-test

echo "✅ Docker gRPC test completed!"