package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Name        string
	Enviornment string
}

type GRPCConfig struct {
	Port             string
	EnableReflection bool
}

type HTTPCOnfig struct {
	Port string
}

type MongoDbConfig struct {
	URI      string
	Database string
}

type JWTCOnfig struct {
	Secret              string
	Issuer              string
	AccessTokenDuration time.Duration
}

type Config struct {
	App     AppConfig
	GRPC    GRPCConfig
	HTTP    HTTPCOnfig
	MongoDb MongoDbConfig
	JWT     JWTCOnfig
}

func (c *Config) Validate() error {
	if c.MongoDb.URI == "" {
		return fmt.Errorf("MONGO_DB URI is required")
	}

	if c.MongoDb.Database == "" {
		return fmt.Errorf("Database name is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}

	if c.JWT.Issuer == "" {
		return fmt.Errorf("jwt issuer is required")
	}

	if c.JWT.AccessTokenDuration <= 0 {
		return fmt.Errorf("token life must be greater than 0")
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

	accessTokenDuration, err := getEnvDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "user-service"),
			Enviornment: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPCOnfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
		GRPC: GRPCConfig{
			Port:             getEnv("GRPC_PORT", "50051"),
			EnableReflection: getEnvBool("GRPC_REFLECTION_ENABLED", true),
		},
		MongoDb: MongoDbConfig{
			URI:      getEnv("MONGO_URI", os.Getenv("MOMGO_URI")),
			Database: os.Getenv("MONGO_DATABASE"),
		},
		JWT: JWTCOnfig{
			Secret:              os.Getenv("JWT_SECRET"),
			Issuer:              getEnv("JWT_ISSUER", "user-service"),
			AccessTokenDuration: accessTokenDuration,
		},
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, err
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("Invalid Duration for %s: %w", key, err)
	}

	return duration, nil
}
