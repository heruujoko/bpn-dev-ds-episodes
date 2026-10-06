# Cache Demo

Interactive Go console app for the "Design System Episodes" cache talk. Five scenarios show when cache helps, when it hurts, how key design affects memory, and what changes with more than one server.

## Prerequisites

Docker (with Compose). Nothing else.

## Run

```bash
cd episodes/cache
docker compose run --rm demo     # builds the demo, starts Redis, opens the menu
docker compose down              # clean up when finished
```

Set `NO_COLOR=1` to disable ANSI colors.

Without Docker: with Go 1.26+ installed, `go run .` works for scenarios 1-4 (`go test ./...` runs the tests). Scenario 5 needs a Redis at `REDIS_ADDR` (default `localhost:6379`); without it, it prints a hint and returns to the menu.

## Scenarios

1. **Cache makes things fast** - 5 reads of a 500ms "DB" call, with and without cache. Shows per-request latency, hit/miss stats, total time.
2. **Anti-pattern: caching permissions** - a 1-hour TTL keeps a revoked `write` permission alive (security bug).
3. **Anti-pattern: caching inventory** - two buyers read a stale cached stock of 1 and both buy; stock ends at -1. The race is made deterministic (both read before either writes), so it reproduces every run.
4. **Cache key granularity** - per-field vs per-user vs profile/contact keys. Memory is estimated from the real key/value sizes of one stored user, scaled to 10,000 users (plus an assumed per-entry overhead). Hit rates come from a seeded 1,000-op simulation (20% email updates). Treat all numbers as estimates, not Redis measurements.
5. **Multiple servers: in-memory vs Redis** - two simulated servers share one fake DB. With a per-server in-memory cache, an update via B leaves A serving stale "Alice". With one shared Redis, the same server code serves fresh "Alicia". Also prints hit latency (in-memory vs Redis) and the trade-offs Redis brings. Needs Redis (included in the compose file).

## Learning objectives

- Cache avoids repeated expensive work; it is a speed vs freshness trade-off.
- Do not cache security-critical or fast-changing, business-critical data with long TTLs.
- Key granularity trades memory, hit rate, and invalidation cost.
- A per-process cache can't be invalidated from another server; a shared cache fixes that and costs a network hop and a new thing that can be down.

## Presentation tips

- Run order: 1 (about 6s) -> 2 -> 3 -> 4 -> 5 (about 2s of simulated DB reads); roughly 1-2 minutes each with talking.
- Scenarios 2 and 3 are instant; pause on the red lines.
- In scenario 4, point out that the finest keys cost the most memory and the coarsest keys invalidate the most.
- Menu: Enter returns to the menu; `0` or Ctrl-D exits.
