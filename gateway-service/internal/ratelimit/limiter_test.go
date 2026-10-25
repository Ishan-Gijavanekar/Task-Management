package ratelimit

import (
	"testing"
	"time"
)

func TestFixedWindowLimitAndReset(t *testing.T) {
	l := New()
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	for i := 0; i < 2; i++ {
		if ok, _ := l.Allow("key", 2, time.Minute); !ok {
			t.Fatal("request denied early")
		}
	}
	if ok, retry := l.Allow("key", 2, time.Minute); ok || retry != time.Minute {
		t.Fatalf("ok=%v retry=%v", ok, retry)
	}
	now = now.Add(time.Minute)
	if ok, _ := l.Allow("key", 2, time.Minute); !ok {
		t.Fatal("window did not reset")
	}
}
func TestKeysAreIsolated(t *testing.T) {
	l := New()
	if ok, _ := l.Allow("a", 1, time.Minute); !ok {
		t.Fatal()
	}
	if ok, _ := l.Allow("b", 1, time.Minute); !ok {
		t.Fatal()
	}
	if ok, _ := l.Allow("a", 1, time.Minute); ok {
		t.Fatal("expected a to be limited")
	}
}
