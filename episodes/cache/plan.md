# Cache Demo — Implementation Plan

## Overview

Implemented interactive Go console application with 5 cache scenarios demonstrating patterns, anti-patterns, design trade-offs, and how the design evolves from per-process memory to shared Redis.

## Task Breakdown

### Task 1: Create Project Structure (5 min)

**Files:**
- `episodes/cache/main.go`
- `episodes/cache/cache/memory.go`
- `episodes/cache/scenarios/happy_path.go`
- `episodes/cache/scenarios/permissions.go`
- `episodes/cache/scenarios/inventory.go`
- `episodes/cache/scenarios/key_granularity.go`
- `episodes/cache/utils/display.go`
- `episodes/cache/scenarios/multi_server.go`
- `episodes/cache/Dockerfile`
- `episodes/cache/docker-compose.yml`
- `episodes/cache/README.md`

**Action:**
- Create directory structure
- Initialize Go module: `go mod init cache-demo`
- Create skeleton files with package declarations

**Verify:**
- `go build` succeeds (empty main)

---

### Task 2: Implement Core Cache (20 min)

**Files:**
- `episodes/cache/cache/memory.go`

**Action:**
1. Define `Cache` struct with mutex, items map, stats
2. Implement `NewCache()` constructor
3. Implement `Get(key string) (interface{}, bool)`:
   - Lock mutex
   - Check if key exists
   - Check if expired (compare ExpiresAt with time.Now())
   - If valid: increment hits, return value + true
   - If expired/missing: increment misses, return nil + false
4. Implement `Set(key string, value interface{}, ttl time.Duration)`:
   - Lock mutex
   - Calculate ExpiresAt = time.Now().Add(ttl)
   - Store in map
5. Implement `Delete(key string)`:
   - Lock mutex
   - Remove from map
6. Implement `Clear()`:
   - Lock mutex
   - Reset items map
7. Implement `Stats() (hits, misses int64, ratio float64)`:
   - Return stats and calculate hit ratio
8. Implement `MemoryUsage() int64`:
   - Estimate memory: count items, approximate size per item
   - Return total bytes

**Verify:**
- Write unit tests in `cache/memory_test.go`
- Test Get/Set with valid TTL
- Test Get with expired TTL
- Test hit/miss counting
- Test Delete and Clear

---

### Task 3: Implement Display Utilities (10 min)

**Files:**
- `episodes/cache/utils/display.go`

**Action:**
1. Create helper functions:
   - `PrintHeader(title string)` — print formatted section header
   - `PrintSuccess(msg string)` — print with ✓ prefix
   - `PrintError(msg string)` — print with ✗ prefix
   - `PrintWarning(msg string)` — print with ⚠ prefix
   - `PrintCritical(msg string)` — print with ❌ prefix
   - `PrintStats(hits, misses int64)` — format statistics
   - `PrintSeparator()` — print visual separator
2. Use ANSI colors for better visibility (optional, check if terminal supports)

**Verify:**
- Test each function prints correctly
- Visual check in terminal

---

### Task 4: Implement Happy Path Scenario (15 min)

**Files:**
- `episodes/cache/scenarios/happy_path.go`

**Action:**
1. Define `HappyPathScenario` struct implementing `Scenario` interface
2. Implement `Name()` and `Description()`
3. Implement `Run()`:
   - Print header: "Scenario 1: Cache Makes Things Fast"
   - Simulate slow DB call: `time.Sleep(500 * time.Millisecond)`
   - Run 5 requests WITHOUT cache:
     - Measure latency per request
     - Print latency + "cache miss"
     - Sum total time
   - Print separator
   - Create new cache
   - Run 5 requests WITH cache:
     - Check cache first
     - If miss: call slow DB, store in cache
     - Measure latency per request
     - Print latency + hit/miss indicator
     - Sum total time
   - Print separator
   - Print cache stats (hits, misses, ratio)
   - Print total time comparison

**Verify:**
- Run scenario manually
- Verify output matches design doc example
- Check timing is realistic (~500ms for miss, <5ms for hit)

---

### Task 5: Implement Permissions Anti-Pattern (15 min)

**Files:**
- `episodes/cache/scenarios/permissions.go`

**Action:**
1. Define `PermissionsScenario` struct
2. Implement `Run()`:
   - Print header: "Scenario 2: Anti-Pattern - Caching Permissions"
   - Step 1: User logs in
     - Fetch permissions from "DB": ["read", "write"]
     - Cache with 1-hour TTL
     - Print cached permissions
   - Step 2: Admin revokes "write" permission
     - Update "DB": ["read"]
     - Print DB updated
     - Print warning: "Cache NOT invalidated!"
   - Step 3: User makes request
     - Fetch from cache
     - Print cached permissions: ["read", "write"]
     - Print critical: "Security breach! User has 'write' access they shouldn't have!"
   - Print lesson: "Don't cache security-critical data with long TTL"

**Verify:**
- Run scenario manually
- Verify security bug is clearly visible
- Check output is dramatic enough for presentation

---

### Task 6: Implement Inventory Anti-Pattern (15 min)

**Files:**
- `episodes/cache/scenarios/inventory.go`

**Action:**
1. Define `InventoryScenario` struct
2. Implement `Run()`:
   - Print header: "Scenario 3: Anti-Pattern - Caching Inventory"
   - Setup: Stock = 1, cache with 10s TTL
   - Simulate concurrent requests (use goroutines + WaitGroup):
     - Request A: check cache → sees 1 → buy → print success
     - Request B: check cache → sees 1 → buy → print success
   - Wait for both to complete
   - Print final stock: -1
   - Print critical: "OVERSOLD! Both requests succeeded because cache showed stale data"
   - Print lesson: "Don't cache frequently-changing data with business consequences"

**Verify:**
- Run scenario manually
- Verify both requests succeed (bug is visible)
- Check final stock shows -1

---

### Task 7: Implement Key Granularity Scenario (20 min)

**Files:**
- `episodes/cache/scenarios/key_granularity.go`

**Action:**
1. Define `KeyGranularityScenario` struct
2. Implement `Run()`:
   - Print header: "Scenario 4: Cache Key Granularity"
   - Strategy 1: Too Granular
     - Create 5 keys per user (name, email, avatar, phone, address)
     - Simulate 10,000 users (just calculate, don't actually create all)
     - Calculate: 50,000 keys, ~5MB memory
     - Print stats
   - Strategy 2: Too Coarse
     - Create 1 key per user (entire object)
     - Calculate: 10,000 keys, ~2MB memory
     - Print stats
     - Show invalidation problem: update email → invalidate entire user
   - Strategy 3: Balanced
     - Create 2 keys per user (profile, contact)
     - Calculate: 20,000 keys, ~3MB memory
     - Print stats
     - Show benefit: update email → invalidate only profile
   - Print recommendation: Strategy 3 balances memory, hit rate, complexity

**Verify:**
- Run scenario manually
- Verify memory calculations are reasonable
- Check output clearly compares strategies

---

### Task 8: Implement Interactive Menu (10 min)

**Files:**
- `episodes/cache/main.go`

**Action:**
1. Import all scenarios
2. Create scenarios slice
3. Implement main loop:
   - Clear screen (optional)
   - Print menu:
     ```
     === Cache Design System Episodes ===
     
     1. Cache makes things fast (happy path)
     2. Anti-pattern: Caching permissions
     3. Anti-pattern: Caching inventory
     4. Cache key granularity comparison
     0. Exit
     
     Choose scenario (0-4):
     ```
   - Read user input
   - Validate choice
   - Run selected scenario
   - Wait for user to press Enter before returning to menu
4. Handle exit (choice = 0)

**Verify:**
- Run `go run main.go`
- Test each menu option
- Verify scenarios run correctly
- Test exit option

---

### Task 9: Create README (5 min)

**Files:**
- `episodes/cache/README.md`

**Action:**
1. Write title and description
2. Add prerequisites (Go 1.21+)
3. Add run instructions:
   ```bash
   cd episodes/cache
   go run main.go
   ```
4. List scenarios with brief descriptions
5. Add learning objectives
6. Add presentation tips (which scenarios to show, timing)

**Verify:**
- README is clear and complete
- Instructions work as written

---

### Task 10: End-to-End Testing (10 min)

**Action:**
1. Run `go build` — verify no errors
2. Run `go run main.go`
3. Test each scenario:
   - Scenario 1: Verify timing and stats
   - Scenario 2: Verify security bug visible
   - Scenario 3: Verify oversell bug visible
   - Scenario 4: Verify memory comparison clear
4. Test menu navigation
5. Test exit
6. Fix any issues found

**Verify:**
- All scenarios run without errors
- Output is clear and presentation-ready
- Code is well-commented

---

## Execution Strategy

**Delegate to code-executor agent:**
- Tasks 1-10 in sequence
- Single agent handles all implementation
- Agent writes tests for cache implementation
- Agent runs end-to-end tests before reporting complete

**Model:** Use medium-tier model (Sonnet/Defenders/X-Men) for execution

**Review:** After execution, review code quality and test coverage

---

## Acceptance Criteria

- [x] `go run .` launches the interactive menu
- [x] All 5 scenarios are implemented
- [x] Scenario 1 shows clear performance improvement
- [x] Scenario 2 shows security bug visibly
- [x] Scenario 3 shows deterministic oversell visibly
- [x] Scenario 4 shows labelled memory/hit-rate estimates
- [x] Scenario 5 shows why per-server memory goes stale and how shared Redis invalidation resolves it
- [x] Output uses visual indicators (✓ ✗ ⚠ ❌)
- [x] Cache and multi-server logic have unit tests
- [x] README makes Docker the only participant prerequisite
- [ ] Pre-talk Docker smoke test: `docker compose run --rm demo`, then run scenario 5 against Redis

---

## Out of Scope

- LRU eviction
- Cache stampede prevention
- Production-grade Redis outage, fallback, and retry policy
- Benchmarking a real production dataset
