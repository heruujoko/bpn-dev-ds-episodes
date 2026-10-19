# Circuit Breaker — Episode 4

## Concept first: what do closed, open and half-open mean? (2 minutes)

Start with the name, because it trips people up. In everyday language "open" sounds good, like an open shop. In an electrical circuit it's the opposite:

| State | Electrical meaning | Software meaning |
|---|---|---|
| **Closed** | Circuit connected, current flows | Calls go through to the dependency normally |
| **Open** | Circuit broken, current stops | Calls are blocked; fail fast without calling the dependency |
| **Half-open** | (software-only state) | Allow one test call to check whether the dependency recovered |

Say it in one line:

> “Closed means calls flow. Open means we've stopped calling. Half-open means we're testing whether to start again.”

Point out that **open is a protective decision, not the failure itself**. The dependency is what's broken; an open circuit is our system choosing to stop waiting on it.

Then show the cycle the demos will walk through:

```
closed --(3 consecutive failures)--> open
open   --(cooldown passes, next request)--> half-open
half-open --(test call succeeds)--> closed
half-open --(test call fails)--> open
```

Keep this table on screen (or come back to it) during Scene 4, when B's circuit actually moves through these states.

## Opening: one checkout, three couriers (1 minute)

“We cached repeated reads, limited incoming traffic, and distributed requests. Now our checkout compares shipping quotes from three courier partners. What happens when just one of them goes down?”

Couriers A, B and C are independent dependencies. Display quotes as options, not automatic purchase decisions. These are fictional services simulated locally; no real orders or external courier APIs are involved.

## Scene 1: normal checkout (1 minute)

Menu 1, `scenarios/healthy.go`.

Ask all three couriers concurrently, with a timeout for each call. Collect and display results in A/B/C order so the console is readable. All three shipping options are available.

The checkout waits for results or deadlines; it is not a streaming UI. Show actual call counts. Any timing displayed must be measured and labelled as simulated work, not a production benchmark.

## Scene 2: courier B is down (2 minutes)

Menu 2, `scenarios/no_breaker.go`.

B stops responding; A and C still quote successfully. Run four checkouts. Every checkout attempts B again and waits for its deadline.

Expected calls: A=4, B=4, C=4. Each checkout still offers A and C, but waits for B's timeout before displaying the complete result.

“A timeout bounds one wait. It does not remember that the last three attempts failed.”

## Scene 3: the wrong boundary — one shared breaker (2 minutes)

Menu 3, `scenarios/shared.go`.

Put one breaker around the entire quote-aggregation operation. The deliberately wrong policy classifies a checkout as failed if ANY courier fails, even when useful quotes from A and C are returned. After three consecutive incomplete checkouts, the shared circuit opens. Checkout four calls nobody: A=3, B=3, C=3.

“No shipping quotes available: shared circuit is open.”

This is not a property of every shared breaker: it is the result of this explicitly chosen operation boundary and failure rule. A/C successes do not reset this breaker because it observes whole-checkout outcomes, not individual calls.

“One partner failed. Our policy disabled all three. The code works exactly as specified—the specification is the mistake.”

## Scene 4: isolate each courier, then recover (4 minutes)

Menu 4, `scenarios/isolated.go`.

Give each courier its own breaker. Three consecutive timeouts open B's circuit; A and C remain closed. Checkout four skips B immediately while A and C still supply quotes.

Expected initial calls: A=4, B=3, C=4.

Customer message: “Quotes available from Courier A and Courier C. Courier B is temporarily unavailable.” Do not invent a zero-price quote or silently change a customer's selected courier. If there are no successful quotes, explicitly report that no shipping option is available.

### Recovery is tested, not assumed

Advance an injected breaker clock by the cooldown (no long wait on stage):

1. B is still down. Admit one half-open probe; its failure reopens B's circuit.
2. An immediate subsequent checkout skips B again.
3. Mark B healthy and advance another cooldown. One successful probe closes B's circuit.
4. Next checkout returns all three quotes normally.

Print the transitions: **closed → open → half-open → open → half-open → closed**. Point back to the concept table from the start: this is the same cycle, now happening to Courier B. While the half-open test call is in flight, other calls to B are still rejected.

## Closing: choose the failure boundary (1 minute)

“One courier going down should remove one shipping option—not take down checkout.”

A circuit breaker does not repair the courier. It temporarily avoids calls that are likely to fail, reducing wasted waits and attempts. Calls already in flight still need timeouts; a breaker is not a concurrency limit or a retry policy.

Thresholds and cooldowns are demo settings, not universal best practices. Invalid addresses are business rejections, not evidence that a courier is unavailable. Production failure classification must follow the partner's API contract.

Unlike a global rate-limit quota, breaker state can reasonably remain local to each caller and dependency. Multiple instances may probe separately; centralizing every breaker in Redis is not an automatic improvement.

## Demo limits

- Three concurrent simulated courier calls in one Go process; no HTTP or real payments.
- Fixed failure threshold of 3 and cooldown of 30 simulated seconds.
- Separate example files preserve the healthy, no-breaker, and wrong-boundary baselines.
- Retries, stale-quote fallbacks, distributed breaker coordination and real HTTP containers are not part of this episode.
