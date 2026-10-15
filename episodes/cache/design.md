# Cache Demo — Design Document

## Overview

Interactive Go console application demonstrating cache patterns, anti-patterns, and design trade-offs for the "Design System Episodes" talk.

## Architecture

### Single Binary, Multiple Scenarios

```
episodes/cache/
├── main.go                    # Entry point and interactive menu
├── cache/memory.go            # In-memory cache, TTL, stats, size estimate
├── scenarios/
│   ├── happy_path.go          # Slow query with/without cache
│   ├── permissions.go         # Anti-pattern: cached permissions
│   ├── inventory.go           # Anti-pattern: cached inventory
│   ├── key_granularity.go     # Cache-key design comparison
│   └── multi_server.go        # Per-server memory vs shared Redis
├── utils/display.go           # Console formatting helpers
├── Dockerfile                 # Static binary in a scratch image
└── docker-compose.yml         # Demo + Redis, Docker-only setup
```

### Core Components

#### 1. In-Memory Cache (`cache/memory.go`)

```go
type Cache struct {
    mu      sync.RWMutex
    items   map[string]*CacheItem
    stats   CacheStats
}

type CacheItem struct {
    Value     interface{}
    ExpiresAt time.Time
}

type CacheStats struct {
    Hits   int64
    Misses int64
}
```

**Features:**
- Thread-safe (mutex)
- TTL-based expiration
- Hit/miss tracking
- Memory usage calculation

**Methods:**
- `Get(key string) (interface{}, bool)` — returns value + hit/miss
- `Set(key string, value interface{}, ttl time.Duration)` — stores with TTL
- `Delete(key string)` — explicit invalidation
- `Clear()` — flush all
- `Stats() CacheStats` — returns hit/miss counts
- `MemoryUsage() int64` — estimated bytes

#### 2. Scenario Runner Pattern

Each scenario implements:

```go
type Scenario interface {
    Name() string
    Description() string
    Run() error
}
```

#### 3. Interactive Menu (`main.go`)

```go
func main() {
    scenarios := []Scenario{
        &HappyPathScenario{},
        &PermissionsScenario{},
        &InventoryScenario{},
        &KeyGranularityScenario{},
        &MultiServerScenario{},
    }
    
    for {
        displayMenu(scenarios)
        choice := readChoice()
        if choice == 0 { break }
        scenarios[choice-1].Run()
    }
}
```

## Scenario Designs

### Scenario 1: Happy Path

**Without cache:**
```go
for i := 0; i < 5; i++ {
    start := time.Now()
    user := fetchUserFromDB(123)  // simulates 500ms DB call
    fmt.Printf("Request %d: %v (no cache)\n", i+1, time.Since(start))
}
```

**With cache:**
```go
cache := NewCache()
for i := 0; i < 5; i++ {
    start := time.Now()
    user, hit := cache.Get("user:123")
    if !hit {
        user = fetchUserFromDB(123)
        cache.Set("user:123", user, time.Minute)
    }
    fmt.Printf("Request %d: %v (hit=%v)\n", i+1, time.Since(start), hit)
}
```

**Output:**
```
=== Scenario 1: Cache Makes Things Fast ===

Without cache:
Request 1: 500ms (cache miss)
Request 2: 500ms (cache miss)
Request 3: 500ms (cache miss)
Request 4: 500ms (cache miss)
Request 5: 500ms (cache miss)
Total: 2500ms

With cache:
Request 1: 500ms (cache miss, stored)
Request 2: 2ms   ✓ (cache hit)
Request 3: 2ms   ✓ (cache hit)
Request 4: 2ms   ✓ (cache hit)
Request 5: 2ms   ✓ (cache hit)
Total: 508ms

Stats: 4 hits, 1 miss, 80% hit ratio
```

### Scenario 2: Permissions Anti-Pattern

**Setup:**
- User has permissions: ["read", "write"]
- Cache with 1-hour TTL
- Admin revokes "write" permission
- User still has "write" because cache hasn't expired

**Flow:**
```go
// 1. User logs in, permissions cached
cache.Set("user:123:perms", []string{"read", "write"}, time.Hour)

// 2. Admin revokes permission in DB
db.RevokePermission(123, "write")

// 3. User makes request
perms, _ := cache.Get("user:123:perms")
// Returns ["read", "write"] ❌ STALE!
```

**Output:**
```
=== Scenario 2: Anti-Pattern - Caching Permissions ===

Step 1: User logs in
Permissions cached: [read, write]
TTL: 1 hour

Step 2: Admin revokes 'write' permission
Database updated: [read]
Cache NOT invalidated!

Step 3: User makes request
Cache returns: [read, write] ⚠ STALE!
Security breach: User has 'write' access they shouldn't have!

Lesson: Don't cache security-critical data with long TTL.
```

### Scenario 3: Inventory Anti-Pattern

**Setup:**
- Product has 1 item in stock
- Cache stock count
- Two concurrent requests buy the item
- Both succeed because cache shows stale count

**Flow:**
```go
// Initial state
cache.Set("product:456:stock", 1, time.Second*10)

// Concurrent requests
go func() {
    stock, _ := cache.Get("product:456:stock")
    if stock > 0 {
        buy()  // succeeds
    }
}()

go func() {
    stock, _ := cache.Get("product:456:stock")
    if stock > 0 {
        buy()  // also succeeds! ❌
    }
}()
```

**Output:**
```
=== Scenario 3: Anti-Pattern - Caching Inventory ===

Initial state:
Stock: 1
Cache: 1 (TTL: 10s)

Request A: checks stock → cache says 1 → buys ✓
Request B: checks stock → cache says 1 → buys ✓

Result:
Stock: -1 ❌ OVERSOLD!

Lesson: Don't cache frequently-changing data with business consequences.
```

### Scenario 4: Key Granularity

**Three strategies:**

```go
// Strategy 1: Too granular
cache.Set("user:123:name", "Alice", time.Hour)
cache.Set("user:123:email", "alice@example.com", time.Hour)
cache.Set("user:123:avatar", "url", time.Hour)
cache.Set("user:123:phone", "555-1234", time.Hour)
cache.Set("user:123:address", "123 Main St", time.Hour)

// Strategy 2: Too coarse
cache.Set("user:123", entireUserObject, time.Hour)

// Strategy 3: Balanced
cache.Set("user:123:profile", profileObject, time.Hour)
cache.Set("user:123:contact", contactObject, time.Hour)
```

**Output:**
```
=== Scenario 4: Cache Key Granularity ===

Strategy 1: Too Granular
Keys: user:123:name, user:123:email, user:123:avatar, user:123:phone, user:123:address
For 10,000 users: 50,000 keys
Memory: ~5.0 MB
Hit rate: 95% (fine-grained updates)

Strategy 2: Too Coarse
Keys: user:123
For 10,000 users: 10,000 keys
Memory: ~2.0 MB
Hit rate: 60% (entire object invalidated on any change)

Strategy 3: Balanced
Keys: user:123:profile, user:123:contact
For 10,000 users: 20,000 keys
Memory: ~3.0 MB
Hit rate: 85% (logical grouping)

Recommendation: Strategy 3 balances memory, hit rate, and invalidation complexity.
```

## Implementation Plan

### Phase 1: Core Cache (30 min)
1. Create `cache/memory.go` with Get/Set/Delete/Clear
2. Add TTL expiration logic
3. Add hit/miss tracking
4. Add memory usage estimation
5. Write unit tests

### Phase 2: Scenarios
1. Implement `HappyPathScenario`
2. Implement `PermissionsScenario`
3. Implement `InventoryScenario`
4. Implement `KeyGranularityScenario`
5. Implement `MultiServerScenario`: the same server logic first uses two separate in-memory stores (stale read), then a shared Redis store (fresh read)

### Phase 3: Docker-only delivery
1. Build a static binary with Go's alpine image and run it from `scratch`
2. Run Redis from its official alpine image with a healthcheck
3. Configure Compose so `docker compose run --rm demo` waits for Redis and opens the interactive menu

### Phase 4: UI, test, and documentation
1. Add console formatting helpers and visual indicators (✓ ✗ ⚠ ❌)
2. Unit-test cache operations and both multi-server outcomes using in-memory stores
3. Add README run instructions and a pre-talk Docker smoke-test command

## Acceptance Criteria

- [x] Menu displays all 5 scenarios
- [x] Scenarios 1–4 run without dependencies beyond Go
- [x] Scenario 5 fails safely with a clear instruction when Redis is unavailable
- [x] The per-server-memory flow visibly produces a stale read; the shared-store flow has unit-test coverage for a fresh read
- [x] Output clearly shows cache hits/misses and anti-pattern bugs
- [x] Key granularity shows labelled estimates, not fabricated Redis measurements
- [x] Docker Compose is the documented participant path; its image uses a static binary and `scratch` runtime
- [ ] Pre-talk smoke test: run `docker compose run --rm demo`, choose scenario 5, and confirm the shared Redis path live

## Out of Scope

- LRU eviction (too complex for this demo)
- Cache stampede prevention
- Performance benchmarks beyond the illustrative scenario timings
- Production-grade Redis outage, retry, or fallback policy

## Future Enhancements

- Add cache stampede demonstration
- Add cache warming scenario
- Add visualization (charts/graphs)
