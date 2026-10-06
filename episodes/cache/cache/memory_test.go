package cache

import (
	"testing"
	"time"
)

func TestGetSet(t *testing.T) {
	c := NewCache()
	c.Set("k", "v", time.Minute)

	v, ok := c.Get("k")
	if !ok || v != "v" {
		t.Fatalf("Get(k) = %v, %v; want v, true", v, ok)
	}
}

func TestGetMissing(t *testing.T) {
	c := NewCache()
	if v, ok := c.Get("nope"); ok || v != nil {
		t.Fatalf("Get(nope) = %v, %v; want nil, false", v, ok)
	}
}

func TestExpiry(t *testing.T) {
	c := NewCache()
	c.Set("k", "v", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if _, ok := c.Get("k"); ok {
		t.Fatal("expired entry should be a miss")
	}
}

func TestHitMissCounts(t *testing.T) {
	c := NewCache()
	c.Set("k", 1, time.Minute)
	c.Get("k")    // hit
	c.Get("k")    // hit
	c.Get("nope") // miss

	hits, misses, ratio := c.Stats()
	if hits != 2 || misses != 1 {
		t.Fatalf("hits=%d misses=%d; want 2, 1", hits, misses)
	}
	if want := 2.0 / 3.0; ratio < want-1e-9 || ratio > want+1e-9 {
		t.Fatalf("ratio = %v; want %v", ratio, want)
	}
}

func TestExpiredCountsAsMiss(t *testing.T) {
	c := NewCache()
	c.Set("k", 1, 5*time.Millisecond)
	time.Sleep(15 * time.Millisecond)
	c.Get("k")

	if hits, misses, _ := c.Stats(); hits != 0 || misses != 1 {
		t.Fatalf("hits=%d misses=%d; want 0, 1", hits, misses)
	}
}

func TestStatsZeroRatio(t *testing.T) {
	// No lookups yet: ratio must be 0, not NaN.
	if _, _, ratio := NewCache().Stats(); ratio != 0 {
		t.Fatalf("ratio = %v; want 0", ratio)
	}
}

func TestDelete(t *testing.T) {
	c := NewCache()
	c.Set("k", "v", time.Minute)
	c.Delete("k")
	if _, ok := c.Get("k"); ok {
		t.Fatal("deleted key should miss")
	}
	c.Delete("never-existed") // must not panic
}

func TestClear(t *testing.T) {
	c := NewCache()
	c.Set("a", 1, time.Minute)
	c.Set("b", 2, time.Minute)
	c.Clear()

	if _, ok := c.Get("a"); ok {
		t.Fatal("Clear should remove a")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("Clear should remove b")
	}
	if c.Len() != 0 {
		t.Fatalf("Len = %d; want 0", c.Len())
	}
}

func TestMemoryUsageGrowsWithData(t *testing.T) {
	c := NewCache()
	if c.MemoryUsage() != 0 {
		t.Fatal("empty cache should use 0 bytes")
	}
	c.Set("a", "short", time.Minute)
	small := c.MemoryUsage()
	c.Set("b", "a much longer value than the first one", time.Minute)
	if c.MemoryUsage() <= small {
		t.Fatal("memory usage should grow as items are added")
	}
}

func TestEstimateSize(t *testing.T) {
	// 16-byte string header + 5 bytes of data.
	if got := EstimateSize("hello"); got != 16+5 {
		t.Fatalf("EstimateSize(string) = %d", got)
	}
	if got := EstimateSize([]string{"ab", "c"}); got != 24+(16+2)+(16+1) {
		t.Fatalf("EstimateSize([]string) = %d", got)
	}
	if got := EstimateSize(42); got != 8 {
		t.Fatalf("EstimateSize(int) = %d", got)
	}
}
