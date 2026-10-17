# Rate Limit Episode — Storyboard

> Episode 2: Who gets to use the system, how much, and what happens when they ask for more.

## Opening Hook (2 min)

**Transition from Episode 1:**
"Cache made each request cheaper. But cheaper is not free. What happens when one client sends 10,000 requests per second? Or an AI agent gets stuck in a retry loop against your API?"

**Problem statement:**
Every system has a finite budget: CPU, DB connections, third-party API quota, money. Without a limit, the noisiest client decides how that budget is spent, and everyone else pays.

**Thesis:**
Writing a rate limiter is easy. AI will generate one in seconds. Deciding **who** to limit, **how much**, **by which key**, and **what the user sees** when limited is the part that needs a human.

---

## Scene 1: No Limit — One Noisy Client Ruins It for Everyone (3 min)

**Setup:**
- A backend that can handle only a few requests at a time (simulated capacity)
- A bot floods it with requests
- Alice, a normal user, sends a few requests at the same time

**Demo (illustrative output):**
```
Backend capacity: 5 concurrent requests

Bot:   sends 50 requests
Alice: sends 3 requests

Alice request 1: waited 2.1s  ⚠
Alice request 2: waited 2.3s  ⚠
Alice request 3: TIMEOUT      ❌

Bot consumed 94% of backend capacity.
```

**Key takeaway:**
Without a limit, capacity goes to whoever asks fastest, not to whoever needs it.

---

## Scene 2: Fixed Window — Simple, and Mostly Right (3 min)

**The idea:**
"Each client gets N requests per window." Count requests per key; reset the counter when the window ends.

**Demo:**
```
Limit: 5 requests per 1s window, per client

Bot request 1-5:  ✓ allowed
Bot request 6-50: ✗ 429 Too Many Requests (Retry-After: 1s)
Alice request 1-3: ✓ allowed, fast
```

**The weakness — boundary burst:**
```
t=0.9s  Bot sends 5  → ✓ allowed (window 1)
t=1.0s  window resets
t=1.1s  Bot sends 5  → ✓ allowed (window 2)

10 requests in 0.2s, while the "limit" says 5 per second. ⚠
```

**Key takeaway:**
Fixed window is cheap and easy to reason about. It allows up to 2× the limit around window edges. Often fine, sometimes not.

---

## Scene 3: Token Bucket — Allow Bursts, Enforce the Average (4 min)

**The idea:**
A bucket holds up to N tokens. Each request spends one. Tokens refill at a steady rate. Empty bucket → 429.

**Demo:**
```
Bucket: capacity 5, refill 1 token/second

t=0.0s  burst of 5    → ✓✓✓✓✓ (bucket empty)
t=0.1s  request 6     → ✗ 429 (Retry-After: 1s)
t=1.0s  request 7     → ✓ (1 token refilled)
t=1.1s  request 8     → ✗ 429
```

**Side by side with fixed window at the boundary:**
```
Fixed window: 10 requests allowed in 0.2s  ⚠
Token bucket:  5 requests allowed in 0.2s  ✓
```

**Key takeaway:**
Real users are bursty (page load = many requests at once). Token bucket allows a burst up to the capacity while holding the long-run average. It is usually the practical default.

---

## Scene 4: The Key Is the Policy — Who Are You Limiting? (4 min)

**Problem:**
A rate limit is `limit[key]`. Choosing the key is choosing who shares a budget.

### Anti-pattern 1: Limit by IP only

```
Office network: 20 employees behind one NAT IP
Limit: 10 requests/min per IP

Employee 1-10:  ✓ allowed
Employee 11-20: ✗ 429 ❌ legitimate users blocked
```

**Lesson:** IPs are shared (offices, mobile carriers, campuses). An IP limit punishes the innocent people standing next to the noisy one.

### Anti-pattern 2: Limit by user ID only, on the login endpoint

```
Attacker is NOT logged in → there is no user ID
Limit by user ID → no key → no limit

Attacker tries 1,000 passwords for alice@example.com: all allowed ❌
```

**Lesson:** Unauthenticated endpoints need a different key, e.g. the **target account** (`login:alice@example.com`) plus the IP.

### Balanced

| Endpoint | Key | Why |
|---|---|---|
| Authenticated API | user ID / API key | fair per customer, not per network |
| Login | target account + IP | stops password guessing on one account |
| Public signup / search | IP (generous limit) | the only identity available |

**Key takeaway:**
The algorithm matters less than the key. The wrong key either blocks real users or protects nothing.

---

## Scene 5: The Plot Twist — What If We Have Multiple Servers? (4 min)

**Transition question (callback to Episode 1):**
"Remember the cache plot twist? Our rate limiter keeps its counters in memory. Now we run 3 servers behind a load balancer. The limit says 5. How many requests does the bot actually get?"

**Part 1 — per-server in-memory counters:**
```
Limit: 5 per window, 3 servers, round-robin

Server A counter: 5  ✓✓✓✓✓
Server B counter: 5  ✓✓✓✓✓
Server C counter: 5  ✓✓✓✓✓

Bot got 15 requests through. Limit was 5. ❌
```

**Part 2 — evolve: shared counter in Redis:**
```
All servers INCR the same key: ratelimit:bot:<window>

Requests 1-5:   ✓ allowed
Requests 6-15:  ✗ 429

Bot got 5 requests through. ✓
```

**The new trade-offs Redis brings:**
- A network round trip on every request, before any real work happens
- Redis is down → do you **fail open** (allow everything, stay available) or **fail closed** (block everything, stay protected)? This is a business decision, not a technical one.
- Counter and expiry must be set atomically, or a crash can leave a key that never expires (the user is limited forever)

**Key takeaway:**
Same lesson as cache: state that lived in one process has to move to a shared place once there are multiple servers, and that shared place has a cost.

*(Sets up the Load Balancing episode: we keep saying "behind a load balancer". Next episode opens that box.)*

---

## Scene 6: Interactive Demo — You Decide (2 min)

**Menu-driven scenarios:**
```
1. No limit: one noisy client
2. Fixed window (and its boundary burst)
3. Token bucket
4. Choosing the key: IP vs user vs account
5. Multiple servers: in-memory vs Redis (run via `docker compose run --rm demo`, Redis included)
```

---

## Closing (2 min)

**Summary:**
1. Without a limit, the noisiest client spends everyone's capacity
2. Fixed window is simple; token bucket handles real bursty traffic better
3. The key is the policy: IP, user, account, or API key each protect different things
4. Once you have multiple servers, counters must be shared, and you must choose fail-open or fail-closed
5. A good 429 says *when* to retry (`Retry-After`), so legitimate clients can behave

**The fundamental truth about rate limiting:**

There is no universal correct limit. "100 requests per minute" is not a best practice; it is a guess unless it comes from your capacity and your users' real behavior.

**The core trade-off:**

Every rate limit is a trade-off between **protection** and **access**.

- Too strict → you block paying customers, break page loads, and generate support tickets
- Too loose → one client, bug, or AI agent loop can take the system down for everyone
- The right limit depends on what you are protecting and who your users are

**Context matters:**
- Login endpoint? Protect accounts → strict, keyed by account
- Paid API? Fairness per customer → per API key, tiered by plan
- Public read endpoint? Availability → generous, keyed by IP
- Expensive AI/LLM call? Cost → strict, per user, possibly per day

**The human judgment AI can't replace:**
AI can write a token bucket. It doesn't know that your biggest customer sends all traffic from one NAT IP, that your login page is under credential stuffing, or that blocking checkout costs more than an overloaded server.

**Call to action:**
"Next time you add a rate limit, ask: What am I protecting? Who shares this budget? What does a legitimate user see when they hit it? And what happens when the limiter itself is down?"

---

## Technical Implementation Notes

**Language:** Go (stdlib + go-redis for scenario 5)
**Run:** `docker compose run --rm demo` (Docker is the only prerequisite)

**Key metrics to display:**
- Allowed vs rejected (429) count per client
- Latency / wait time for the normal user
- Requests let through vs the configured limit
- `Retry-After` on every rejection

**Visual indicators:**
- ✓ Allowed
- ✗ Rejected (429)
- ⚠ Degraded / suspicious
- ❌ Bug caused by the limiter design
