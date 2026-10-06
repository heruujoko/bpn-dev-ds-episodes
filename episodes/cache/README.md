# Cache Demo

Interactive Go console app for the "Design System Episodes" cache talk. Four scenarios show when cache helps, when it hurts, and how key design affects memory. Stdlib only, no dependencies.

## Prerequisites

Go 1.21+ (developed on 1.27).

## Run

```bash
cd episodes/cache
go run .          # or: go run main.go
go test ./...     # unit tests for the cache core
```

Set `NO_COLOR=1` to disable ANSI colors.

## Scenarios

1. **Cache makes things fast** - 5 reads of a 500ms "DB" call, with and without cache. Shows per-request latency, hit/miss stats, total time.
2. **Anti-pattern: caching permissions** - a 1-hour TTL keeps a revoked `write` permission alive (security bug).
3. **Anti-pattern: caching inventory** - two buyers read a stale cached stock of 1 and both buy; stock ends at -1. The race is made deterministic (both read before either writes), so it reproduces every run.
4. **Cache key granularity** - per-field vs per-user vs profile/contact keys. Memory is estimated from the real key/value sizes of one stored user, scaled to 10,000 users (plus an assumed per-entry overhead). Hit rates come from a seeded 1,000-op simulation (20% email updates). Treat all numbers as estimates, not Redis measurements.

## Learning objectives

- Cache avoids repeated expensive work; it is a speed vs freshness trade-off.
- Do not cache security-critical or fast-changing, business-critical data with long TTLs.
- Key granularity trades memory, hit rate, and invalidation cost.

## Presentation tips

- Run order: 1 (about 6s) -> 2 -> 3 -> 4; roughly 1-2 minutes each with talking.
- Scenarios 2 and 3 are instant; pause on the red lines.
- In scenario 4, point out that the finest keys cost the most memory and the coarsest keys invalidate the most.
- Menu: Enter returns to the menu; `0` or Ctrl-D exits.
