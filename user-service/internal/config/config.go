package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Name        string
	Environment string
}

type GRPCConfig struct {
	Port             string
	EnableReflection bool
}

type HTTPConfig struct {
	Port           string
	AllowedOrigins string
	BodyLimitMB    int
	TrustedProxies []string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type JWTConfig struct {
	Secret              string
	ActiveKeyID         string
	PreviousKeys        map[string]string
	Issuer              string
	AccessTokenDuration time.Duration
}

type AuthConfig struct {
	RefreshTokenDuration     time.Duration
	RequireEmailVerification bool
	RefreshCookieName        string
	RefreshCookieSecure      bool
	FrontendURL              string
	SMTPHost                 string
	SMTPPort                 string
	SMTPUsername             string
	SMTPPassword             string
	SMTPFrom                 string
}

type Config struct {
	App     AppConfig
	GRPC    GRPCConfig
	HTTP    HTTPConfig
	MongoDB MongoDBConfig
	JWT     JWTConfig
	Auth    AuthConfig
}

func (c *Config) Validate() error {
	if strings.Contains(c.HTTP.AllowedOrigins, "*") {
		return fmt.Errorf("HTTP_ALLOWED_ORIGINS cannot contain wildcard when credentials are enabled")
	}
	if c.MongoDB.URI == "" {
		return fmt.Errorf("MONGO_DB URI is required")
	}

	if c.MongoDB.Database == "" {
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
	if c.Auth.RefreshTokenDuration <= 0 {
		return fmt.Errorf("refresh token life must be greater than 0")
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
	refreshTokenDuration, err := getEnvDuration("AUTH_REFRESH_TOKEN_DURATION", 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	bodyLimitMB, err := getEnvInt("HTTP_BODY_LIMIT_MB", 1)
	if err != nil {
		return nil, err
	}
	if bodyLimitMB <= 0 {
		return nil, fmt.Errorf("HTTP_BODY_LIMIT_MB must be positive")
	}

	cfg := Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "user-service"),
			Environment: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Port:           getEnv("HTTP_PORT", "8080"),
			AllowedOrigins: getEnv("HTTP_ALLOWED_ORIGINS", "http://localhost:5173"),
			BodyLimitMB:    bodyLimitMB,
			TrustedProxies: getEnvCSV("HTTP_TRUSTED_PROXIES"),
		},
		GRPC: GRPCConfig{
			Port:             getEnv("GRPC_PORT", "50051"),
			EnableReflection: getEnvBool("GRPC_REFLECTION_ENABLED", true),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGO_URI", os.Getenv("MOMGO_URI")),
			Database: os.Getenv("MONGO_DATABASE"),
		},
		JWT: JWTConfig{
			Secret:              os.Getenv("JWT_SECRET"),
			ActiveKeyID:         getEnv("JWT_ACTIVE_KEY_ID", "current"),
			PreviousKeys:        getEnvKeyMap("JWT_PREVIOUS_KEYS"),
			Issuer:              getEnv("JWT_ISSUER", "user-service"),
			AccessTokenDuration: accessTokenDuration,
		},
		Auth: AuthConfig{
			RefreshTokenDuration:     refreshTokenDuration,
			RequireEmailVerification: getEnvBool("AUTH_REQUIRE_EMAIL_VERIFICATION", false),
			RefreshCookieName:        getEnv("AUTH_REFRESH_COOKIE_NAME", "refresh_token"),
			RefreshCookieSecure:      getEnvBool("AUTH_REFRESH_COOKIE_SECURE", true),
			FrontendURL:              getEnv("AUTH_FRONTEND_URL", "http://localhost:5173"),
			SMTPHost:                 os.Getenv("SMTP_HOST"),
			SMTPPort:                 getEnv("SMTP_PORT", "587"),
			SMTPUsername:             os.Getenv("SMTP_USERNAME"),
			SMTPPassword:             os.Getenv("SMTP_PASSWORD"),
			SMTPFrom:                 os.Getenv("SMTP_FROM"),
		},
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, err
}

func getEnvCSV(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnvKeyMap(key string) map[string]string {
	result := make(map[string]string)
	for _, item := range getEnvCSV(key) {
		parts := strings.SplitN(item, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}

	return parsed, nil
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
