package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Name        string
	Enviornment string
}

type GRPCConfig struct {
	Port string
}

type HTTPCOnfig struct {
	Port string
}

type MongoDbConfig struct {
	URI      string
	Database string
}

type Config struct {
	App     AppConfig
	GRPC    GRPCConfig
	HTTP    HTTPCOnfig
	MongoDb MongoDbConfig
}

func (c *Config) Validate() error {
	if c.MongoDb.URI == "" {
		return fmt.Errorf("MONGO_DB URI is required")
	}

	if c.MongoDb.Database == "" {
		return fmt.Errorf("Database name is required")
	}

	return nil
}

func getEnv(name, fallback string) string {
	value := os.Getenv(name)

	if value == "" {
		return fallback
	}

	return value
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "user-service"),
			Enviornment: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPCOnfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "50051"),
		},
		MongoDb: MongoDbConfig{
			URI:      getEnv("MONGO_URI", os.Getenv("MOMGO_URI")),
			Database: os.Getenv("MONGO_DATABASE"),
		},
	}

	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, err
}
