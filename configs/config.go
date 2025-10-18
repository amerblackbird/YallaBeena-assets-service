package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Kafka    KafkaConfig    `json:"kafka"`
	Storage  StorageConfig  `json:"storage"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	GRPCPort  int    `json:"grpc_port"`
	ApiPrefix string `json:"api_prefix"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"ssl_mode"`
}

type StorageConfig struct {
	Endpoint   string `json:"endpoint"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	UseSSL     bool   `json:"use_ssl"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string    `json:"brokers"`
	GroupID string      `json:"group_id"`
	Topics  KafkaTopics `json:"topics"`
}

// KafkaTopics defines all Kafka topics
type KafkaTopics struct {
	ActivityLogs string `json:"activity_logs"`
	UsersEvents  string `json:"users_events"`
	AssetsEvents string `json:"assets_events"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:      getEnv("SERVER_HOST", "localhost"),
			Port:      getEnvAsInt("SERVER_PORT", 8080),
			GRPCPort:  getEnvAsInt("GRPC_PORT", 9090),
			ApiPrefix: getEnv("API_PREFIX", "/api/v1"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "auth_service_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			GroupID: getEnv("KAFKA_GROUP_ID", "assets-service"),
			Topics: KafkaTopics{
				UsersEvents:   getEnv("KAFKA_TOPIC_USERS_EVENTS", "user.events"),
				AssetsEvents: getEnv("KAFKA_TOPIC_ASSETS_EVENTS", "assets.events"),
				ActivityLogs: getEnv("KAFKA_TOPIC_ACTIVITY_LOGS_EVENTS", "activity.logs"),
			},
		},
		Storage: StorageConfig{
			Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
			BucketName: getEnv("MINIO_BUCKET_NAME", "assets"),
			Region:     getEnv("MINIO_REGION", "us-east-1"),
			UseSSL:     getEnvAsBool("MINIO_USE_SSL", false),
		},
	}

	return config, nil
}

// DatabaseURL returns the database connection URL
func (c *DatabaseConfig) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// RedisURL returns the Redis connection URL
func (c *RedisConfig) RedisURL() string {
	if c.Password != "" {
		return fmt.Sprintf("%s@%s:%d/%d", c.Password, c.Host, c.Port, c.DB)
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Helper functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}
