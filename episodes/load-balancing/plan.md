# Load Balancing Implementation Plan

**Goal:** Ship only the approved two-example slice on `load-balancing-demo`.
**Architecture:** Deterministic stdlib Go routing functions, separate console scenarios, one Docker Compose demo service. See `design.md` for contracts and `storyboard.md` for narration.
**Tech Stack:** Go 1.26, Go testing, alpine builder, scratch runtime, Docker Compose.

## 1. Save the narrative checkpoint

Files: `storyboard.md`, `design.md`, `plan.md` under `episodes/load-balancing/`.
Commit and push these documents before implementation. No changes to earlier episodes.

## 2. Routing functions and tests (5–8 minutes)

Create `go.mod`, `balancer/balancer.go`, `balancer/balancer_test.go`.
Write failing tests first for round-robin [0,1,2] repeated, weighted [0,0,0,1,1,2] repeated, totals 4/4/4 versus 6/4/2, one backend, zero-weight skipping, invalid input, and weight-sum overflow. Implement the minimum stateless selectors described in design.md.
Verify from the module: `mise exec go -- go test ./...`.

## 3. Console examples (5–8 minutes)

Create `main.go`, `scenarios/round_robin.go`, `scenarios/weighted.go`.
Use the earlier episode menu as the reference; no extra display package unless necessary. Print assignments and computed totals. Keep each scenario in a separate file. Explain simulation, grouped weighting, and equal-count/unequal-work caveats.
Verify menu choices 1, 2, 0, invalid input and EOF. No sleeps or random scheduling.

## 4. Docker and participant instructions (3–5 minutes)

Create `Dockerfile`, `docker-compose.yml`, `.dockerignore`, `README.md`.
Use `golang:1.26-alpine` builder and `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"`; copy only binary to scratch. Use numeric non-root USER. Compose has only `demo`, with stdin and TTY enabled. Stdlib means no go.sum to copy. Docker is the only prerequisite.

## 5. Review, verify, commit, push (reserve 5 minutes)

Main reviews the integrated diff and runs:
- `mise exec go -- go vet ./...`
- `mise exec go -- go test -race -count=1 ./...`
- `docker compose build`
- `printf '1\n\n2\n\n0\n' | docker compose run --rm -T demo`
- `git diff --check`

Commit only verified working files and push this branch. Do not create a PR or merge. At the timebox boundary, stop and record remaining work here; do not publish broken scaffolding.

## Completed checkpoint

- Storyboard/design/plan pushed first in `7c8a8f5`.
- Routing tests initially failed before implementation, then passed with race detection; vet passed.
- Docker build and both menu scenarios passed: 12/0/0 baseline, 4/4/4 round-robin, 6/4/2 weighted.
- Five invalid-input/EOF smoke cases passed. Image reported 2.39 MB on this machine.
- Only the two approved examples implemented. Resume with a separately approved next slice.

## Deferred

Unequal-duration workloads, adaptive routing, health checks, sticky sessions, and actual HTTP backends. These require a separate approved slice. No work on those today.
