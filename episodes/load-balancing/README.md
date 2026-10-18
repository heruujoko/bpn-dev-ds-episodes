# Load Balancing Demo

Episode 3, first slice: two deterministic Go console examples. Docker with Compose is the only participant prerequisite.

## Run

```bash
cd episodes/load-balancing
docker compose run --build --rm demo
# Choose 1 or 2, Enter to return, 0 to exit.
docker compose down
```

Compose runs one container: a stripped static binary in a non-root `scratch` image. No Redis or backend services are needed for this simulation.

## Examples

1. **One server → round-robin** (`scenarios/round_robin.go`): 12 requests all sent to A (12/0/0), followed by routing A/B/C cyclically (4/4/4).
2. **Different capacities → weighted routing** (`scenarios/weighted.go`): compare round-robin with grouped weights 3:2:1. Two cycles produce 6/4/2 assignments.

Each example prints individual assignments and computes totals. Keep both files open during the talk to compare them.

## What this does not prove

The backends are simulated in one process. There are no HTTP requests, latency benchmarks, or throughput measurements. Equal request counts do not guarantee equal work. Grouped weighted routing can send bursts to a backend; it is not smooth weighted round-robin. Weights do not automatically track live capacity or failures.

The callback to Episodes 1–2: routing does not make per-server cache entries or rate-limit counters shared.

## Developer checks (optional Go 1.26+)

Run from this module directory:

```bash
go vet ./...
go test -race ./...
printf '1\n\n2\n\n0\n' | docker compose run --build --rm -T demo
```

For editor navigation, open `episodes/load-balancing` as the workspace; it is an independent Go module like the earlier episodes. No repository-wide Go workspace is required.

## Continue later

The approved slice stops here. Unequal-duration requests, least-active-request routing, health checks, sticky sessions, and real HTTP backends remain deferred. See `storyboard.md`, `design.md`, and `plan.md`.
