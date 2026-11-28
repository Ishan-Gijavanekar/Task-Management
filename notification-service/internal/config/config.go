package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment       string
	HttpPort          string
	GrpcPort          string
	MongoURI          string
	MongoDatabase     string
	NATSURL           string
	NATSstream        string
	StartupTimeout    time.Duration
	DependencyTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func getEnv(key, fallback string) string {
	value, exsists := os.LookupEnv(key)
	if !exsists {
		return fallback
	}
	return value
}

func validatePort(port string) error {
	value, err := strconv.Atoi(port)
	if err != nil || value < 0 || value > 63535 {
		return fmt.Errorf("invalid port number")
	}
	return nil
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, exsists := os.LookupEnv(key)
	if !exsists {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s duration cannot be parsed", key)
	}

	return duration, nil
}

func (c Config) Validate() error {
	switch c.Environment {
	case "development", "test", "production":
	default:
		return fmt.Errorf("invalid APP_ENV")
	}

	if err := validatePort(c.HttpPort); err != nil {
		return fmt.Errorf("HTTP_PORT: %w", err)
	}

	if err := validatePort(c.GrpcPort); err != nil {
		return fmt.Errorf("GRPC_PORT: %w", err)
	}

	if c.HttpPort == c.GrpcPort {
		return fmt.Errorf(
			"HTTP_PORT and GRPC_PORT cannot be identical",
		)
	}

	if strings.TrimSpace(c.MongoURI) == "" {
		return fmt.Errorf("MONGO_URI is required")
	}

	if strings.TrimSpace(c.MongoDatabase) == "" {
		return fmt.Errorf("MONGO_DATABASE is required")
	}

	parsedURL, err := url.Parse(c.NATSURL)
	if err != nil || parsedURL.Host == "" {
		return fmt.Errorf("invalid NATS_URL")
	}

	if parsedURL.Scheme != "nats" &&
		parsedURL.Scheme != "tls" {
		return fmt.Errorf("unsupported NATS_URL scheme")
	}

	if strings.TrimSpace(c.NATSstream) == "" {
		return fmt.Errorf("NATS_STREAM is required")
	}

	if c.StartupTimeout <= 0 ||
		c.DependencyTimeout <= 0 ||
		c.ShutdownTimeout <= 0 {
		return fmt.Errorf("timeouts must be positive")
	}

	return nil
}

func Load() (Config, error) {
	cfg := Config{
		Environment: getEnv("APP_ENV", "development"),

		HttpPort: getEnv("HTTP_PORT", "8082"),
		GrpcPort: getEnv("GRPC_PORT", "50053"),

		MongoURI: getEnv(
			"MONGO_URI",
			"mongodb://localhost:27019/?directConnection=true&replicaSet=rs0",
		),

		MongoDatabase: getEnv(
			"MONGO_DATABASE",
			"notifications",
		),

		NATSURL: getEnv(
			"NATS_URL",
			"nats://localhost:4222",
		),

		NATSstream: getEnv(
			"NATS_STREAM",
			"DOMAIN_EVENTS",
		),
	}

	var err error

	cfg.StartupTimeout, err = getDuration(
		"STARTUP_TIMEOUT",
		30*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	cfg.DependencyTimeout, err = getDuration(
		"DEPENDENCY_TIMEOUT",
		5*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	cfg.ShutdownTimeout, err = getDuration(
		"SHUTDOWN_TIMEOUT",
		15*time.Second,
	)
	if err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
