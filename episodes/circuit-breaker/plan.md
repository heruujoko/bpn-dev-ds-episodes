# Circuit Breaker Implementation Plan

**Goal:** Demonstrate how a failed courier should remove one shipping option, not disable healthy partners.
**Architecture:** One Go module with concurrent simulated couriers, per-call deadlines, injected breaker clock, four separate menu scenarios, and a stdlib-only Docker binary.
**Tech Stack:** Go 1.26, testing, Docker Compose, alpine builder and non-root scratch runtime.

All paths below are relative to `episodes/circuit-breaker/`. Scope and acceptance are in `design.md`; narration is in `storyboard.md`.

## 1. State machine (tests first)

Create `go.mod`, `breaker/breaker.go`, `breaker/breaker_test.go`.
Write tests for threshold 3, reset after success, open rejection without callback, cooldown boundary, half-open success/failure, only one concurrent probe, late old-generation completions, and invalid configuration. Use controlled channels and an injected clock, not sleep-based scheduling. Observe tests failing, then implement the minimal mutex-protected state machine. Run `mise exec go -- go test -race ./breaker`.

## 2. Courier simulation and collection

Create `courier/courier.go`, `courier/courier_test.go` and a small `scenarios` helper.
Model names/prices, call counts, healthy/down behavior and context cancellation. Collect all three quotes concurrently with individual deadlines and stable result ordering. Test successful quotes and down calls returning on cancellation; make errors visible instead of swallowing them. Do not add a real HTTP server.

## 3. Four separate demonstrations

Create `scenarios/healthy.go`, `scenarios/no_breaker.go`, `scenarios/shared.go`, `scenarios/isolated.go`, and `scenarios/scenarios_test.go`.
Run the storyboard sequences with fresh state per menu run. Test the call counts 4/4/4, 3/3/3 and 4/3/4, partial availability, and recovery transitions. In the shared scenario, explicitly classify any failed partner as a failed aggregate operation. In isolation, count failures independently per courier.

## 4. Menu and Docker

Create `main.go`, `README.md`, `Dockerfile`, `docker-compose.yml`, `.dockerignore`.
Use Episode 3 as the minimal structure reference. Docker only prerequisite; `docker compose run --build --rm demo`. No ports, Redis, third-party Go packages or additional runtime image libraries. Handle invalid menu input and EOF. README documents simulated clock, local breaker scope, failure classification and production limitations.

## 5. Main-thread review and verification

Review only this episode's changes. Run from the module:

```bash
mise exec go -- go vet ./...
mise exec go -- go test -race -count=1 ./...
docker compose build
printf '1\n\n2\n\n3\n\n4\n\n0\n' | docker compose run --rm -T demo
docker compose down
git diff --check
```

Check actual output for isolation, state transitions and honest partial results; validate invalid input and EOF. Report any skipped check. Do not modify old examples to make new ones fit.

## Delivery boundary

Implement only after plan approval. No push, PR, merge, host package installation, or changes to earlier episodes. Ask before publishing this branch. Further examples (HTTP, retries, stale quote fallbacks, distributed coordination) are deferred.
