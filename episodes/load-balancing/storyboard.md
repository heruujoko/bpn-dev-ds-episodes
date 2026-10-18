# Load Balancing — Episode 3

## Opening: open the box (1 minute)

“For two episodes we have said ‘behind a load balancer.’ Now let's open that box. Cache avoids repeated work. Rate limits control admitted work. How do we distribute the work that remains?”

These examples simulate three backends in one Go process. No HTTP requests or real server performance measurements are involved. Counts are computed from routing decisions, not hard-coded output.

## Scene 1: one server → round-robin (3 minutes)

Ask: “We have 12 requests and three equally capable servers. Where should each request go?”

Show the baseline: all 12 go to A; B and C are idle. Then run round-robin:

```
A → B → C → A → B → C → A → B → C → A → B → C
A: 4 requests   B: 4 requests   C: 4 requests
```

Menu option 1 prints each assignment and the totals. The lesson is distribution, not a claimed speedup: we have not measured concurrency, latency, or throughput.

**Non-happy-path question:** “What if one request costs 1 ms and another costs 1 second?” Equal request counts are not equal work. Round-robin does not inspect queue length or request cost.

## Scene 2: unequal capacity → weighted round-robin (3 minutes)

Suppose our capacity measurements suggest relative weights A=3, B=2, C=1. Use the simple repeating schedule:

```
A → A → A → B → B → C  (repeat twice)
A: 6 requests   B: 4 requests   C: 2 requests
```

Menu option 2 prints the same 12-request round-robin baseline and the weighted assignments. Keep the examples in separate files so they can be compared during the session.

**Trade-offs:**
- Weights are configuration based on assumptions or measurements, not live load sensing.
- This teaching implementation is grouped weighted round-robin, not smooth weighted round-robin. Consecutive assignments can create bursts.
- Exact proportions occur over complete cycles; short runs may differ.
- Routing more work to A does not prove A can process it faster. Validate weights against the real workload.

## Callback to Episodes 1 and 2 (1 minute)

“We changed where requests land, not where state lives.” Per-server caches remain separate, and per-server counters do not become a global quota simply because we balance traffic. Use the rate-limit episode's scaling comparison to revisit that distinction.

## Closing (1 minute)

There is no universally best routing policy. Round-robin is a simple starting point for comparable backends and requests. Weights express known capacity differences, but neither policy here accounts for current load or failures.

Before asking AI to implement a balancer, decide: what should be balanced—requests, active work, or latency? What evidence supports the weights? What happens when a server slows down or disappears?

## Later sessions — not implemented in this slice

- Unequal request duration and least-active-request routing.
- Health checks and removal/recovery of unhealthy backends.
- Sticky sessions and local-state trade-offs.
- Actual HTTP backend containers, then the dependency-failure transition to circuit breakers.
