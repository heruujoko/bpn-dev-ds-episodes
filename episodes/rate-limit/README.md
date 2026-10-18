# Rate Limit Demo

Interactive Go console app for the "Design System Episodes" rate-limit talk. Five scenarios show why limits matter, how fixed windows and token buckets differ, why the key is policy, and what changes with more than one server.

## Prerequisites

Docker (with Compose). Nothing else.

## Run

```bash
cd episodes/rate-limit
docker compose run --rm demo     # builds the demo, starts Redis, opens the menu
docker compose down              # clean up when finished
```

Set `NO_COLOR=1` to disable ANSI colors.

## Scenarios

1. **No limit: one noisy client** - a bot deterministically fills five backend slots before Alice arrives, so her measured waits and timeouts show how capacity is spent.
2. **Fixed window (and its boundary burst)** - five requests per second per client, then ten requests allowed across a window edge.
3. **Token bucket** - allows a burst of five while enforcing one-token-per-second refill afterward.
4. **Choosing the key: IP vs user vs account** - IP blocks office colleagues, empty user IDs bypass login limits, and target-account-plus-IP blocks guessing.
5. **Multiple servers: in-memory vs Redis** - three in-memory counters allow 15 requests through a limit of five; a shared Redis counter allows five.
6. **Scaling out: local counters vs centralized Redis** - routing makes local limits over- or under-enforce; a shared Redis counter stays at five.

## Learning objectives

- Without a limit, the noisiest client spends shared capacity first.
- Fixed windows are simple; token buckets preserve a long-run average while allowing bursts.
- A rate-limit key is policy: it decides who shares a budget.
- Per-process counters multiply with servers; a shared store fixes that but adds network and availability trade-offs.
- Counter expiry must be atomic so one user cannot be limited forever.

## Presentation tips

- Run order: 1 -> 2 -> 3 -> 4 -> 5; all scenarios finish quickly.
- In scenario 1, pause on Alice's measured waits and the bot's completed-capacity share.
- In scenario 4, compare the unfair NAT result to the login bypass before showing the balanced table.
- In scenario 5, emphasize that fail-open versus fail-closed is a business decision.
- Menu: Enter returns to the menu; `0` or Ctrl-D exits.
