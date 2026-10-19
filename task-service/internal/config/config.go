package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPPort, GRPCPort, MongoURI, MongoDatabase, JWTIssuer, JWTActiveKeyID string
	JWTKeys                                                                map[string]string
	ShutdownTimeout, DependencyTimeout                                     time.Duration
	ConcurrencyLimit                                                       int
}

func Load() (*Config, error) {
	shutdown, e := duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if e != nil {
		return nil, e
	}
	dependency, e := duration("DEPENDENCY_TIMEOUT", 2*time.Second)
	if e != nil {
		return nil, e
	}
	limit, e := strconv.Atoi(env("SERVICE_CONCURRENCY_LIMIT", "200"))
	if e != nil || limit < 1 {
		return nil, fmt.Errorf("SERVICE_CONCURRENCY_LIMIT must be positive")
	}
	active := env("JWT_ACTIVE_KEY_ID", "current")
	keys := parseKeys(os.Getenv("JWT_KEYS"))
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		keys[active] = secret
	}
	c := &Config{HTTPPort: env("HTTP_PORT", "8081"), GRPCPort: env("GRPC_PORT", "50052"), MongoURI: os.Getenv("MONGO_URI"), MongoDatabase: env("MONGO_DATABASE", "tasks"), JWTIssuer: env("JWT_ISSUER", "user-service"), JWTActiveKeyID: active, JWTKeys: keys, ShutdownTimeout: shutdown, DependencyTimeout: dependency, ConcurrencyLimit: limit}
	if c.MongoURI == "" || len(c.JWTKeys) == 0 {
		return nil, fmt.Errorf("MONGO_URI and JWT_SECRET or JWT_KEYS are required")
	}
	return c, nil
}
func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
func duration(k string, d time.Duration) (time.Duration, error) {
	v := env(k, d.String())
	parsed, e := time.ParseDuration(v)
	if e != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s", k)
	}
	return parsed, nil
}
func parseKeys(v string) map[string]string {
	out := map[string]string{}
	for _, item := range strings.Split(v, ",") {
		p := strings.SplitN(item, ":", 2)
		if len(p) == 2 && strings.TrimSpace(p[0]) != "" && strings.TrimSpace(p[1]) != "" {
			out[strings.TrimSpace(p[0])] = strings.TrimSpace(p[1])
		}
	}
	return out
}
