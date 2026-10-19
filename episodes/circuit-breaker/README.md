# Circuit Breaker Demo

Episode 4: checkout compares three fictional courier partners. B is down; A and C are healthy. Should one lost shipping option take down checkout?

## Run — Docker only

```bash
cd episodes/circuit-breaker
docker compose run --build --rm demo
# Choose 1–4; Enter returns to the menu; 0 exits.
docker compose down
```

One non-root scratch container, a static Go binary, no Redis, network ports, or external courier APIs. Each example has a separate file under `scenarios/`; earlier episodes are unchanged.

## Demonstrations

| Menu | Example | Calls after four checkouts (A/B/C) |
|---|---|---|
| 1 | All couriers healthy | 4/4/4; three quotes available |
| 2 | B times out without a breaker | 4/4/4; A/C quote, each checkout waits for B |
| 3 | Wrongly scoped shared breaker | 3/3/3; fourth checkout blocks even A/C |
| 4 | Independent per-courier breakers | 4/3/4; fourth checkout skips B, A/C still quote |

Option 4 continues through failed and successful half-open probes. Final call counts are 8/6/8. B's state moves closed → open → half-open → open → half-open → closed; only one probe can be in flight.

The shared-breaker error is intentional: it wraps the entire quote collection and treats ANY missing quote as a failed aggregate operation. It is not a claim that every shared breaker behaves this way.

## Timing and honest results

All three simulated calls run concurrently with individual 60ms deadlines. Healthy partners return immediately; an unresponsive partner waits for cancellation. Checkout waits for all outcomes and prints them in stable order. Measured durations reflect this simulation, not real shipping-service performance.

The breaker's cooldown is 30 **simulated** seconds. An injected clock advances between completed checkouts; you never wait 30 seconds during the demo. Deadlines use real time. Output says unavailable for missing quotes—never zero-price shipping or a pretend booking.

## Checks (optional local Go 1.26+)

```bash
go vet ./...
go test -race -count=1 ./...
printf '1\n\n2\n\n3\n\n4\n\n0\n' | docker compose run --build --rm -T demo
```

Tests cover constructor validation, consecutive failures/reset, open rejection, cooldown boundaries, single-probe concurrency, old-generation completions, context cancellation, scenario call counts, partial availability, recovery and menu EOF/invalid input.

## Teaching limits

This is not a production breaker library. All errors returned to this breaker count as dependency failures; real integrations must classify business rejections (e.g. invalid addresses) separately. No rolling failure-rate windows, retries, panic recovery, persistent state, concurrency limit or cancellation of existing calls is implemented. Callbacks must return normally and respect their deadlines.

Local breaker state per caller/dependency is deliberate. Unlike enforcing a global quota, centralized Redis is not automatically necessary. Multiple callers may probe independently. Thresholds, cooldown, fallback policy and failure boundaries need business-specific judgment.

See `storyboard.md`, `design.md`, and `plan.md` for the narrative and scope.
