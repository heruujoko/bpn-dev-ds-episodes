package limiter

import (
	"testing"
	"time"
)

func TestFixedWindowEnforcesLimitAndResets(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewFixedWindow(2, time.Second, func() time.Time { return now })
	for range 2 {
		if ok, _ := l.Allow("bot"); !ok {
			t.Fatal("request should be allowed")
		}
	}
	if ok, retry := l.Allow("bot"); ok || retry != time.Second {
		t.Fatalf("third request = %v, %v; want false, 1s", ok, retry)
	}
	now = now.Add(time.Second)
	if ok, _ := l.Allow("bot"); !ok {
		t.Fatal("window should reset")
	}
}

func TestBoundaryBurstFixedWindowButNotTokenBucket(t *testing.T) {
	now := time.Unix(0, 900*time.Millisecond.Nanoseconds())
	clock := func() time.Time { return now }
	fixed := NewFixedWindow(5, time.Second, clock)
	bucket := NewTokenBucket(5, 5, clock)

	for range 5 {
		if ok, _ := fixed.Allow("bot"); !ok {
			t.Fatal("fixed first burst blocked")
		}
		if ok, _ := bucket.Allow("bot"); !ok {
			t.Fatal("bucket first burst blocked")
		}
	}
	now = now.Add(200 * time.Millisecond)
	for range 5 {
		if ok, _ := fixed.Allow("bot"); !ok {
			t.Fatal("fixed second burst blocked")
		}
	}
	if ok, _ := bucket.Allow("bot"); !ok {
		t.Fatal("one token should refill after a second")
	}
	if ok, _ := bucket.Allow("bot"); ok {
		t.Fatal("token bucket must not allow a boundary burst")
	}
}

func TestTokenBucketRefills(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewTokenBucket(2, 1, func() time.Time { return now })
	for range 2 {
		l.Allow("alice")
	}
	if ok, retry := l.Allow("alice"); ok || retry != time.Second {
		t.Fatalf("empty bucket = %v, %v; want false, 1s", ok, retry)
	}
	now = now.Add(time.Second)
	if ok, _ := l.Allow("alice"); !ok {
		t.Fatal("refilled token should allow request")
	}
}
