# Design System Episodes

> Bite-size system design concepts in everyday development.

## Why This Matters

The rise of AI coding outsources implementation from manual human typing to agent typing and reviewing code. Without humans fully understanding **what**, **when**, and **how** a design system should be used or chosen, the result will be:

- AI slop
- Unnecessary complexity
- Underperformance
- Less secure implementations

## Topics (ordered easy → complex)

### 1. Cache
- Simple in-memory cache (map-based)?
- Redis-backed cache?
- Cache invalidation strategies (TTL, LRU)?
- Show the difference between cached vs. uncached performance?

### 2. Rate Limit
- API endpoint with rate limiting?
- Show what happens when limit is exceeded?
- Different algorithms (fixed window, sliding window, token bucket)?
- Redis-backed distributed rate limit?

### 3. Load Balancing
- Multiple backend instances (fake services)?
- Different strategies (round-robin, least connections, weighted)?
- Health checks?
- Show traffic distribution visually in console?

### 4. Circuit Breaker
- Call a slow/failing external service?
- Show state transitions (closed → open → half-open)?
- Fallback behavior when circuit is open?
- Configurable thresholds (failure count, timeout)?

## Tech Stack

- **Language:** Go (console-based apps)
- **Infrastructure:** Docker Compose + Docker images
- **Dependencies:** Redis (for cache + rate limit), mock services as needed

## Open Questions

- Demo flow: standalone `go run main.go` per app? Interactive console menu?
- Real-time metrics/stats in console output?
- "Before/after" demonstrations for each pattern?
- Shared utilities package across demos?
- Code structure: separate directories per demo?
