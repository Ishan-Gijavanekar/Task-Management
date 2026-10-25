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
	HTTPPort        string
	PublicOrigin    string
	UserServiceURL  *url.URL
	TaskServiceURL  *url.URL
	JWTIssuer       string
	JWTKeys         map[string]string
	BodyLimit       int
	ResponseLimit   int64
	UpstreamTimeout time.Duration
	ShutdownTimeout time.Duration
	TrustedProxies  []string
}

func Load() (*Config, error) {
	userURL, err := serviceURL("USER_SERVICE_URL", "http://localhost:8080")
	if err != nil {
		return nil, err
	}
	taskURL, err := serviceURL("TASK_SERVICE_URL", "http://localhost:8081")
	if err != nil {
		return nil, err
	}
	timeout, err := duration("UPSTREAM_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	shutdown, err := duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}
	bodyMB, err := positiveInt("HTTP_BODY_LIMIT_MB", 2)
	if err != nil {
		return nil, err
	}
	responseMB, err := positiveInt("HTTP_RESPONSE_LIMIT_MB", 8)
	if err != nil {
		return nil, err
	}
	active := env("JWT_ACTIVE_KEY_ID", "current")
	keys := keyMap(os.Getenv("JWT_KEYS"))
	if secret := strings.TrimSpace(os.Getenv("JWT_SECRET")); secret != "" {
		keys[active] = secret
	}
	cfg := &Config{HTTPPort: env("HTTP_PORT", "8088"), PublicOrigin: env("PUBLIC_ORIGIN", "http://localhost:5173"), UserServiceURL: userURL, TaskServiceURL: taskURL, JWTIssuer: env("JWT_ISSUER", "user-service"), JWTKeys: keys, BodyLimit: bodyMB * 1024 * 1024, ResponseLimit: int64(responseMB) * 1024 * 1024, UpstreamTimeout: timeout, ShutdownTimeout: shutdown, TrustedProxies: csv(os.Getenv("HTTP_TRUSTED_PROXIES"))}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c == nil || c.UserServiceURL == nil || c.TaskServiceURL == nil {
		return fmt.Errorf("service URLs are required")
	}
	if c.UserServiceURL.Scheme == "" || c.UserServiceURL.Host == "" || c.TaskServiceURL.Scheme == "" || c.TaskServiceURL.Host == "" {
		return fmt.Errorf("service URLs must be absolute")
	}
	if c.UserServiceURL.Scheme != "http" && c.UserServiceURL.Scheme != "https" {
		return fmt.Errorf("user service URL scheme must be http or https")
	}
	if c.TaskServiceURL.Scheme != "http" && c.TaskServiceURL.Scheme != "https" {
		return fmt.Errorf("task service URL scheme must be http or https")
	}
	if strings.Contains(c.PublicOrigin, "*") || strings.Contains(c.PublicOrigin, ",") {
		return fmt.Errorf("PUBLIC_ORIGIN must be one explicit origin")
	}
	origin, err := url.Parse(c.PublicOrigin)
	if err != nil || origin.Scheme == "" || origin.Host == "" || origin.Path != "" {
		return fmt.Errorf("PUBLIC_ORIGIN must be an absolute origin")
	}
	if c.JWTIssuer == "" || len(c.JWTKeys) == 0 {
		return fmt.Errorf("JWT_ISSUER and JWT_SECRET or JWT_KEYS are required")
	}
	if c.BodyLimit <= 0 || c.ResponseLimit <= 0 || c.UpstreamTimeout <= 0 || c.ShutdownTimeout <= 0 {
		return fmt.Errorf("limits and timeouts must be positive")
	}
	return nil
}

func serviceURL(key, fallback string) (*url.URL, error) {
	raw := env(key, fallback)
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", key, err)
	}
	return parsed, nil
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func duration(key string, fallback time.Duration) (time.Duration, error) {
	value, err := time.ParseDuration(env(key, fallback.String()))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}
func positiveInt(key string, fallback int) (int, error) {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return value, nil
}
func csv(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
func keyMap(raw string) map[string]string {
	out := map[string]string{}
	for _, item := range csv(raw) {
		parts := strings.SplitN(item, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
			out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return out
}
