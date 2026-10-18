// Package limiter provides small deterministic rate-limiting algorithms for the demo.
package limiter

import (
	"sync"
	"time"
)

// FixedWindow allows limit requests per window for each key.
type FixedWindow struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	counters map[string]fixedCounter
}

type fixedCounter struct {
	start time.Time
	count int
}

// NewFixedWindow creates a limiter whose clock can be injected for deterministic demos and tests.
func NewFixedWindow(limit int, window time.Duration, now func() time.Time) *FixedWindow {
	return &FixedWindow{limit: limit, window: window, now: now, counters: make(map[string]fixedCounter)}
}

// Allow reports whether key may proceed and, when rejected, how long to wait for its window to reset.
func (l *FixedWindow) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	n := l.now()
	c := l.counters[key]
	windowStart := n.Truncate(l.window)
	if c.start != windowStart {
		c = fixedCounter{start: windowStart}
	}
	if c.count >= l.limit {
		return false, l.window - n.Sub(c.start)
	}
	c.count++
	l.counters[key] = c
	return true, 0
}

// TokenBucket allows bursts up to capacity and refills tokens steadily.
type TokenBucket struct {
	mu       sync.Mutex
	capacity float64
	refill   float64 // tokens per second
	now      func() time.Time
	buckets  map[string]bucket
}

type bucket struct {
	tokens float64
	at     time.Time
}

// NewTokenBucket creates a limiter whose clock can be injected for deterministic demos and tests.
func NewTokenBucket(capacity int, refillPerSecond float64, now func() time.Time) *TokenBucket {
	return &TokenBucket{capacity: float64(capacity), refill: refillPerSecond, now: now, buckets: make(map[string]bucket)}
}

// Allow reports whether key may proceed and, when rejected, the time until one token refills.
func (l *TokenBucket) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	n := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = bucket{tokens: l.capacity, at: n}
	} else {
		b.tokens = min(l.capacity, b.tokens+n.Sub(b.at).Seconds()*l.refill)
		b.at = n
	}
	if b.tokens >= 1 {
		b.tokens--
		l.buckets[key] = b
		return true, 0
	}
	l.buckets[key] = b
	if l.refill <= 0 {
		return false, 0
	}
	return false, time.Duration((1 - b.tokens) / l.refill * float64(time.Second))
}
