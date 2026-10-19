# Circuit Breaker Demo Design

## Approved story

Checkout requests quotes from three independent couriers. B becomes unresponsive while A/C stay healthy. Compare no breaker, a wrongly scoped aggregation breaker, and isolated per-courier breakers with recovery. Four separate example files; standard library only; same Docker console workflow as earlier episodes.

## Components

- `breaker/breaker.go`: concurrency-safe closed/open/half-open state machine, injected clock, configurable positive failure threshold and cooldown, and `Execute(func() error) error`. An explicit open-circuit error distinguishes rejected work from dependency failure.
- `courier/courier.go`: fictional partner name, integer price, healthy/down mode, call count. Quotes take a context; healthy calls return a fictional quote, down calls wait for context cancellation. Each admitted call increments the count. Protect mutable state with a mutex.
- `scenarios/`: `healthy.go`, `no_breaker.go`, `shared.go`, `isolated.go`, plus only the small common quote-collection/display helper needed by these examples.
- `main.go`: menu, validation, clean EOF/exit; no unnecessary framework.
- `Dockerfile`/Compose: one static non-root scratch container, Go alpine builder. No Redis or external backend containers.

## Breaker correctness

Closed admits calls. A success resets consecutive failures; a failure increments them. At threshold, open until the cooldown expires. At or after expiry, the next caller atomically reserves the sole half-open probe; concurrent callers are rejected. Probe success closes/reset counts; probe failure reopens and starts a new cooldown.

Do not hold the mutex while executing dependency work. Use a generation/epoch to ignore completions admitted before a state transition: a late success from an older closed generation must not close a newly opened circuit. Validate constructor arguments. State inspection must not admit probes or mutate transitions.

The demo's dependency calls have per-call deadlines, so probes finish. A small teaching breaker need not implement retries, cancellation of existing calls, fallback execution, rolling percentages, or panic recovery; document those limits rather than claiming production readiness.

## Quote flow and policy

Launch A/B/C calls concurrently, each with a short real context deadline (e.g. 60 ms). Wait for all results, print them in stable order. No goroutine may outlive quote collection. Use short deadlines to keep the demo fast, but never assert exact wall-clock latency in tests.

The shared-breaker example wraps one complete collection operation and reports an aggregation failure when ANY partner fails. Partial A/C quotes can still be displayed for the first three checkouts. Once this outer circuit opens, no partner gets called.

The isolated example wraps each courier call with that courier's breaker. Only the unresponsive courier fails in this scripted story; invalid-address classification is discussed, not modelled as a fifth scenario. On circuit rejection, show unavailable, not a fabricated price or successful booking.

Advance a fake clock only between fully completed checkout rounds, avoiding data races. The clock controls breaker cooldown only; actual context deadlines still bound down calls. Display transitions by observing state before/inside/after execution or by a minimal result log; keep state logic out of display code.

## Acceptance

- Four healthy/no-breaker checkouts make four calls per partner; no-breaker B attempts all time out, A/C quotes survive.
- Four shared-breaker checkouts make exactly three calls per partner; the fourth is rejected before any courier call.
- Four isolated-breaker checkouts make A=4, B=3, C=4 calls; only B opens.
- Failed half-open probe reopens; successful probe after another cooldown closes; subsequent call reaches B.
- Tests cover one concurrent probe, threshold/reset, cooldown boundary, stale completion, constructor validation, and scenario call counts.
- Docker-only menu smoke test, vet and race tests pass. No changes to Episodes 1–3.
