package config

import (
	"strings"
	"testing"
	"time"
)

func TestGetEnvInt(t *testing.T) {
	t.Setenv("TEST_INTEGER", "42")
	got, err := getEnvInt("TEST_INTEGER", 1)
	if err != nil || got != 42 {
		t.Fatalf("getEnvInt() = %d, %v; want 42, nil", got, err)
	}
}

func TestLoadRejectsInvalidBodyLimit(t *testing.T) {
	t.Setenv("HTTP_BODY_LIMIT_MB", "0")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DATABASE", "test")
	t.Setenv("JWT_SECRET", "secret")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "must be positive") {
		t.Fatalf("Load() error = %v, want positive body-limit error", err)
	}
}

func TestGetEnvDurationRejectsInvalidValue(t *testing.T) {
	t.Setenv("TEST_DURATION", "later")
	if _, err := getEnvDuration("TEST_DURATION", time.Minute); err == nil {
		t.Fatal("getEnvDuration() accepted invalid duration")
	}
}

func TestConfigurationParsesRotationKeysAndTrustedProxies(t *testing.T) {
	t.Setenv("TEST_KEYS", "old:old-secret,older:older-secret")
	keys := getEnvKeyMap("TEST_KEYS")
	if keys["old"] != "old-secret" || keys["older"] != "older-secret" {
		t.Fatalf("getEnvKeyMap() = %#v", keys)
	}
	t.Setenv("TEST_PROXIES", "10.0.0.1, 10.0.0.2")
	proxies := getEnvCSV("TEST_PROXIES")
	if len(proxies) != 2 || proxies[1] != "10.0.0.2" {
		t.Fatalf("getEnvCSV() = %#v", proxies)
	}
}
