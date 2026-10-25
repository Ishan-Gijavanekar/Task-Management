package ratelimit

import (
	"sync"
	"time"
)

type window struct {
	count    int
	resetsAt time.Time
}
type Limiter struct {
	mu         sync.Mutex
	now        func() time.Time
	entries    map[string]window
	operations uint64
}

func New() *Limiter { return &Limiter{now: time.Now, entries: map[string]window{}} }
func (l *Limiter) Allow(key string, limit int, duration time.Duration) (bool, time.Duration) {
	if limit <= 0 || duration <= 0 {
		return false, duration
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	entry, ok := l.entries[key]
	if !ok || !now.Before(entry.resetsAt) {
		entry = window{resetsAt: now.Add(duration)}
	}
	entry.count++
	l.entries[key] = entry
	l.operations++
	if l.operations%1000 == 0 {
		for existing, value := range l.entries {
			if !now.Before(value.resetsAt) {
				delete(l.entries, existing)
			}
		}
	}
	remaining := entry.resetsAt.Sub(now)
	if remaining < 0 {
		remaining = 0
	}
	return entry.count <= limit, remaining
}
