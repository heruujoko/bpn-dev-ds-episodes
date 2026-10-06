package scenarios

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"cache-demo/cache"
	"cache-demo/utils"
)

// MultiServerScenario: two servers, one DB. Per-process cache goes stale; shared Redis does not.
type MultiServerScenario struct{}

func (MultiServerScenario) Name() string { return "Multiple servers: in-memory vs Redis" }
func (MultiServerScenario) Description() string {
	return "Two servers, one DB: stale per-server cache vs shared Redis"
}

// store is the only thing that differs between part 1 and part 2.
type store interface {
	Get(key string) (string, bool)
	Set(key, val string, ttl time.Duration)
	Delete(key string)
}

// memStore adapts the in-process cache.Cache.
type memStore struct{ c *cache.Cache }

func (m memStore) Get(key string) (string, bool) {
	v, ok := m.c.Get(key)
	if !ok {
		return "", false
	}
	return v.(string), true
}
func (m memStore) Set(key, val string, ttl time.Duration) { m.c.Set(key, val, ttl) }
func (m memStore) Delete(key string)                      { m.c.Delete(key) }

// redisStore keeps values in Redis, shared by every server that points at it.
// ponytail: Redis errors are treated as a miss / no-op (no fallback or retry); add a real policy when Redis outages matter.
type redisStore struct{ c *redis.Client }

const redisOpTimeout = time.Second

func (r redisStore) Get(key string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	v, err := r.c.Get(ctx, key).Result()
	return v, err == nil
}
func (r redisStore) Set(key, val string, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	r.c.Set(ctx, key, val, ttl)
}
func (r redisStore) Delete(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	r.c.Del(ctx, key)
}

// fakeDB is the one database both servers share.
type fakeDB struct {
	mu    sync.Mutex
	name  string
	delay time.Duration // simulated query time
}

func (d *fakeDB) get() string {
	time.Sleep(d.delay)
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.name
}

func (d *fakeDB) set(n string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.name = n
}

const profileKey = "demo:user:1:name"

// server is IDENTICAL in both parts; only the store it is given changes.
type server struct {
	name  string
	cache store
	db    *fakeDB
}

// Read returns the profile name and whether it came from the cache.
func (s *server) Read() (string, bool) {
	if v, ok := s.cache.Get(profileKey); ok {
		return v, true
	}
	v := s.db.get()
	s.cache.Set(profileKey, v, time.Hour)
	return v, false
}

// UpdateName writes the DB, then invalidates the cache this server can reach.
func (s *server) UpdateName(n string) {
	s.db.set(n)
	s.cache.Delete(profileKey)
}

func describe(name string, hit bool) string {
	if hit {
		return fmt.Sprintf("HIT, returns %q", name)
	}
	return fmt.Sprintf("miss, loads %q from DB", name)
}

// timedRead reads via s and reports how long it took.
func timedRead(s *server) (string, bool, time.Duration) {
	start := time.Now()
	n, hit := s.Read()
	return n, hit, time.Since(start)
}

// playFlow runs: read via A, update via B, read via A. It returns what the last read saw.
func playFlow(a, b *server) (name string, hit bool, d time.Duration) {
	n, h := a.Read()
	fmt.Printf("1. Read via %s   -> %s\n", a.name, describe(n, h))
	b.UpdateName("Alicia")
	fmt.Printf("2. Update via %s -> DB = %q, %s invalidates the cache it can reach\n", b.name, "Alicia", b.name)
	name, hit, d = timedRead(a)
	fmt.Printf("3. Read via %s   -> %s\n", a.name, describe(name, hit))
	return name, hit, d
}

func verdict(name string) {
	if name != "Alicia" {
		utils.PrintCritical(fmt.Sprintf("STALE! Server A says %q but the DB says \"Alicia\"", name))
	} else {
		utils.PrintSuccess("Fresh: Server A sees \"Alicia\"")
	}
}

func (MultiServerScenario) Run() error {
	utils.PrintHeader("Scenario 5: Multiple Servers - In-Memory vs Redis")

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	newClient := func() *redis.Client { return redis.NewClient(&redis.Options{Addr: addr}) }
	// Plain TCP check first: a down Redis fails fast and without go-redis retry log noise.
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		probe := newClient()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = probe.Ping(ctx).Err()
		cancel()
		probe.Close()
	}
	if err != nil {
		utils.PrintWarning(fmt.Sprintf("Redis is not reachable at %s (%v)", addr, err))
		fmt.Println("This scenario needs Redis. Start it with: docker compose run --rm demo")
		return nil
	}

	const dbDelay = 500 * time.Millisecond

	// Part 1: each server has its OWN in-memory cache.
	fmt.Println("Part 1: each server has its OWN in-memory cache")
	fmt.Println()
	db := &fakeDB{name: "Alice", delay: dbDelay}
	a := &server{"A", memStore{cache.NewCache()}, db}
	b := &server{"B", memStore{cache.NewCache()}, db}
	name, _, memHit := playFlow(a, b)
	verdict(name)
	utils.PrintWarning("B can't reach A's memory. Each server also warms its own cache (N servers = N cold starts, N copies in RAM).")
	fmt.Println()
	utils.PrintSeparator()

	// Part 2: same server code, one shared Redis.
	fmt.Println("Part 2: same code, both servers share ONE Redis")
	fmt.Println()
	ca, cb := newClient(), newClient()
	defer ca.Close()
	defer cb.Close()
	shared := redisStore{ca}
	shared.Delete(profileKey) // clean slate so reruns are repeatable
	defer redisStore{cb}.Delete(profileKey)
	db = &fakeDB{name: "Alice", delay: dbDelay}
	a = &server{"A", shared, db}
	b = &server{"B", redisStore{cb}, db}
	name, _, _ = playFlow(a, b)
	verdict(name)
	_, _, redisHit := timedRead(a) // step 3 was a miss; this one is a real Redis hit
	fmt.Println()
	utils.PrintSeparator()

	fmt.Printf("Latency of a cache hit: in-memory %v, Redis %v (DB read: %v)\n",
		memHit.Round(time.Microsecond), redisHit.Round(time.Microsecond), dbDelay)
	fmt.Println()
	fmt.Println("Same code, same invalidation call. Only WHERE the cache lives changed.")
	fmt.Println("But Redis is not free:")
	fmt.Println("  - a network hop on every hit (small, but not ~0)")
	fmt.Println("  - a new moving part that can be down")
	fmt.Println("  - serialization cost, and a new place to get key design wrong")
	fmt.Println("  - stale reads still possible between the DB write and the delete")
	return nil
}
