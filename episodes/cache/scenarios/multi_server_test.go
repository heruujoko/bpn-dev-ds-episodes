package scenarios

import (
	"os"
	"testing"

	"github.com/redis/go-redis/v9"

	"cache-demo/cache"
)

func servers(sa, sb store) (*server, *server) {
	db := &fakeDB{name: "Alice"}
	return &server{"A", sa, db}, &server{"B", sb, db}
}

func TestSeparateMemoryCachesGoStale(t *testing.T) {
	a, b := servers(memStore{cache.NewCache()}, memStore{cache.NewCache()})
	name, hit, _ := playFlow(a, b)
	if name != "Alice" || !hit {
		t.Fatalf("want stale hit \"Alice\", got %q hit=%v", name, hit)
	}
}

func TestSharedStoreIsFresh(t *testing.T) {
	shared := memStore{cache.NewCache()} // stands in for Redis: one store both servers see
	a, b := servers(shared, shared)
	name, hit, _ := playFlow(a, b)
	if name != "Alicia" || hit {
		t.Fatalf("want fresh miss \"Alicia\", got %q hit=%v", name, hit)
	}
}

func TestSharedRedisIsFresh(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set")
	}
	c := redis.NewClient(&redis.Options{Addr: addr})
	defer c.Close()
	s := redisStore{c}
	s.Delete(profileKey)
	defer s.Delete(profileKey)
	a, b := servers(s, s)
	if name, _, _ := playFlow(a, b); name != "Alicia" {
		t.Fatalf("want fresh \"Alicia\", got %q", name)
	}
}
