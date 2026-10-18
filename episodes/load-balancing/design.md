# Load Balancing Demo Design

## Approved slice

Two independent Go console examples: round-robin and grouped weighted round-robin. Follow the existing episodes' menu and Docker-only participant workflow without modifying them. Standard library only; no Redis, HTTP servers, new infrastructure, or shared cross-episode framework.

## Structure

- `main.go`: two-option menu with exit, invalid-input handling, and clean EOF behavior.
- `balancer/`: pure routing functions with tests. Accept backend count or integer weights and a zero-based request index; return backend index or error. Stateless index-based selection keeps the demo deterministic.
- `scenarios/round_robin.go`: one-server baseline and 12 requests across A/B/C.
- `scenarios/weighted.go`: equal-distribution baseline and 12 requests with weights 3:2:1.
- `Dockerfile`, `docker-compose.yml`, `.dockerignore`: one demo service, alpine Go build stage, static stripped binary, non-root scratch runtime. No network service or published port is necessary.
- `README.md`: Docker-only run instructions, expected counts, simulation limitations.

## Routing contract

Round-robin chooses `requestIndex % backendCount`. Reject nonpositive backend count and negative indexes.

Weighted routing chooses the cumulative-weight interval containing `requestIndex % sum(weights)`. With [3,2,1], each cycle is A,A,A,B,B,C. Zero weight disables a backend. Reject negative weights, empty/all-zero weights, negative indexes, and integer overflow in the sum. Do not expand weights into a large repeated array.

Only fixed local demo data enters these functions, but validation makes their constraints explicit. Errors reach the menu instead of panicking. No production concurrency or health management is implied.

## Observable acceptance

1. Option 1 shows single-server totals 12/0/0 and round-robin totals 4/4/4, derived from assignments.
2. Option 2 shows round-robin totals 4/4/4 and weighted totals 6/4/2 for 12 requests.
3. Repeated runs produce the same routing; tests cover sequences, wraparound, zero weights, and invalid inputs.
4. `docker compose run --build --rm demo` opens the menu using Docker only.
5. Vet, race tests, and Docker menu smoke test pass before publishing implementation commits.

## Deliberate limits

This is a scheduling simulation, not a production proxy. No latency or throughput claims; no automatic retries, least-connections logic, health probes, or sticky sessions. Add those as separate examples later rather than rewriting these baselines.
