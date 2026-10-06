# Cache Episode — Storyboard

> Episode 1: Understanding when, what, and how to cache.

## Opening Hook (2 min)

**Problem statement:**
"Your app is slow. You add cache. It's fast. Great! But wait... why are users seeing stale data? Why is memory usage exploding? Why did we just create a security bug?"

**Thesis:**
Cache is not just "make things fast." It's a design decision with trade-offs. AI can implement cache, but humans must decide what, when, and how.

---

## Scene 1: The Happy Path — Cache Makes Things Fast (3 min)

**Setup:**
- Simulate slow database query (500ms delay)
- Fetch user profile on every request

**Without cache:**
```
Request 1: 500ms (cache miss, hit DB)
Request 2: 500ms (cache miss, hit DB)
Request 3: 500ms (cache miss, hit DB)
Total: 1500ms
```

**With cache:**
```
Request 1: 500ms (cache miss, hit DB, store in cache)
Request 2: 2ms   (cache hit)
Request 3: 2ms   (cache hit)
Total: 504ms
```

**Demo mechanics:**
- Console shows latency per request
- Show hit/miss counter
- Show total time comparison

**Key takeaway:**
Cache is about avoiding repeated work. Identify expensive operations that return the same result.

---

## Scene 2: What NOT to Cache — The Dark Side (5 min)

### Anti-pattern 1: Caching user permissions

**Scenario:**
- Cache user permissions for 1 hour
- Admin revokes user's access
- User still has access for 59 minutes because permissions are cached

**Demo:**
```
1. User logs in, permissions cached: ["read", "write"]
2. Admin revokes "write" permission in DB
3. User makes request → cache returns ["read", "write"] ❌
4. Security breach!
```

**Lesson:**
Don't cache security-critical data with long TTL. Or implement cache invalidation on permission changes.

### Anti-pattern 2: Caching real-time inventory

**Scenario:**
- E-commerce site caches product stock count
- Product has 1 item left
- Two users buy at the same time
- Cache shows 1 item for both → oversell!

**Demo:**
```
1. Stock = 1, cached
2. User A buys → cache still shows 1 (not updated yet)
3. User B buys → cache still shows 1
4. Both succeed → stock goes to -1 ❌
```

**Lesson:**
Don't cache data that changes frequently and has business consequences if stale.

### Anti-pattern 3: Caching one-time reads

**Scenario:**
- Fetch user's last login timestamp
- Cache it for 1 hour
- User never reads it again

**Demo:**
```
Cache entry created: user:123:last_login
Memory used: 100 bytes
Read count: 1
Value: never used again
```

**Lesson:**
Cache only data you'll read multiple times. One-time reads waste memory.

---

## Scene 3: Cache Key Granularity — The Hidden Cost (5 min)

**Problem:**
Cache is key-value. How granular should your keys be?

### Example 1: Too granular

```
Keys created:
- user:123:name
- user:123:email
- user:123:avatar
- user:123:phone
- user:123:address

For 10,000 users: 50,000 keys
Memory: ~5MB
```

**Problem:** Memory bloat. Redis overhead per key adds up.

### Example 2: Too coarse

```
Keys created:
- user:123 (entire user object)

For 10,000 users: 10,000 keys
Memory: ~2MB
```

**Problem:** Update email → invalidate entire user cache. High invalidation cost.

### Example 3: Balanced

```
Keys created:
- user:123:profile (name, email, avatar)
- user:123:contact (phone, address)

For 10,000 users: 20,000 keys
Memory: ~3MB
```

**Benefit:** Update email → invalidate only profile. Reasonable memory.

**Demo mechanics:**
- Show memory counter as we create keys with different strategies
- Show key count
- Show cache hit rates for different granularities

**Key takeaway:**
Cache key design is a trade-off between memory, hit rate, and invalidation complexity. Consider your infrastructure limits (Redis memory) and access patterns.

---

## Scene 4: Interactive Demo — You Decide (3 min)

**Menu-driven scenarios:**

```
1. Run slow query without cache
2. Run slow query with cache
3. Try caching user permissions (see the bug!)
4. Try caching inventory (see the oversell!)
5. Compare cache key strategies (see memory impact)
6. Show hit/miss statistics
```

Let audience choose which scenario to run live.

---

## Closing (2 min)

**Summary:**
1. Cache makes things fast — but only if you cache the right thing
2. Don't cache: security data, real-time data, one-time reads
3. Cache key granularity affects memory and performance
4. AI can implement cache, but humans must make design decisions

**Call to action:**
"Next time you add cache, ask: What am I caching? Why? For how long? What happens if it's stale?"

---

## Technical Implementation Notes

**Language:** Go
**Dependencies:** 
- In-memory cache (map + mutex + TTL)
- Optional: Redis for distributed cache demo

**Demo app structure:**
```go
// Interactive menu
func main() {
    for {
        fmt.Println("1. Slow query without cache")
        fmt.Println("2. Slow query with cache")
        fmt.Println("3. Cache permissions (anti-pattern)")
        fmt.Println("4. Cache inventory (anti-pattern)")
        fmt.Println("5. Compare key strategies")
        fmt.Println("6. Show statistics")
        fmt.Println("0. Exit")
        
        // Run selected scenario
    }
}
```

**Key metrics to display:**
- Request latency (ms)
- Cache hit/miss count
- Hit ratio (%)
- Memory usage (bytes/KB)
- Key count

**Visual indicators:**
- ✓ Cache hit
- ✗ Cache miss
- ⚠ Stale data
- ❌ Bug caused by caching
