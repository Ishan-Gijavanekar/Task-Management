package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {

	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "8082")
	t.Setenv("GRPC_PORT", "50053")

	t.Setenv(
		"MONGO_URI",
		"mongodb://localhost:27019",
	)

	t.Setenv("MONGO_DATABASE", "notifications")
	t.Setenv("NATS_URL", "nats://localhost:4222")
	t.Setenv("NATS_STREAM", "DOMAIN_EVENTS")

	t.Setenv("STARTUP_TIMEOUT", "30s")
	t.Setenv("DEPENDENCY_TIMEOUT", "5s")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")

	cfg, err := Load()

	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}

	if cfg.HttpPort != "8082" {
		t.Errorf("unexpected HTTP port")
	}

	if cfg.MongoDatabase != "notifications" {
		t.Errorf("unexpected database")
	}

	if cfg.DependencyTimeout != 5*time.Second {
		t.Errorf("unexpected dependency timeout")
	}
}

func TestValidateInvalidPort(t *testing.T) {
	cfg := Config{
		HttpPort:    "invalid",
		Environment: "test",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestLoadInvalidTimeout(t *testing.T) {
	t.Setenv("STARTUP_TIMEOUT", "invalid")

	_, err := Load()

	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
}

func TestLoadNegativeTimeout(t *testing.T) {
	t.Setenv("STARTUP_TIMEOUT", "-5s")

	_, err := Load()

	if err == nil {
		t.Fatal("expected negative timeout error")
	}
}

func TestValidateDuplicatePorts(t *testing.T) {
	cfg := Config{
		Environment: "test",
		HttpPort:    "8082",
		GrpcPort:    "8082",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate port error")
	}
}

func TestValidateInvalidNATSURL(t *testing.T) {
	cfg := Config{
		Environment: "test",
		HttpPort:    "8082",
		GrpcPort:    "50053",

		MongoURI:      "mongodb://localhost:27019",
		MongoDatabase: "notifications",

		NATSURL:    "invalid-url",
		NATSstream: "DOMAIN_EVENTS",

		StartupTimeout:    30 * time.Second,
		DependencyTimeout: 5 * time.Second,
		ShutdownTimeout:   15 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid NATS URL error")
	}
}
