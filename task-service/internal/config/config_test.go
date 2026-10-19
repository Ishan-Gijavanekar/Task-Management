package config

import "testing"

func TestLoadRejectsMissingRuntimeSecrets(t *testing.T) {
	t.Setenv("MONGO_URI", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_KEYS", "")
	if _, e := Load(); e == nil {
		t.Fatal("expected validation error")
	}
}
func TestLoadAcceptsRotatingKeys(t *testing.T) {
	t.Setenv("MONGO_URI", "mongodb://localhost")
	t.Setenv("JWT_KEYS", "current:new,old:previous")
	t.Setenv("JWT_SECRET", "")
	c, e := Load()
	if e != nil {
		t.Fatal(e)
	}
	if len(c.JWTKeys) != 2 {
		t.Fatalf("got %d keys", len(c.JWTKeys))
	}
}
