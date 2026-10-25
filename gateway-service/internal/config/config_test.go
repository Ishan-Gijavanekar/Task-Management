package config

import (
	"testing"
	"time"
)

func TestLoadAndValidate(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("PUBLIC_ORIGIN", "https://app.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UserServiceURL.String() != "http://localhost:8080" || cfg.BodyLimit != 2*1024*1024 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestRejectsUnsafeConfiguration(t *testing.T) {
	tests := []Config{
		{PublicOrigin: "*", JWTIssuer: "issuer", JWTKeys: map[string]string{"k": "v"}, BodyLimit: 1, ResponseLimit: 1, UpstreamTimeout: time.Second, ShutdownTimeout: time.Second},
		{PublicOrigin: "https://app.example.com", JWTIssuer: "issuer", JWTKeys: map[string]string{}, BodyLimit: 1, ResponseLimit: 1, UpstreamTimeout: time.Second, ShutdownTimeout: time.Second},
	}
	for i := range tests {
		if tests[i].Validate() == nil {
			t.Errorf("case %d should fail", i)
		}
	}
}
