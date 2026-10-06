package scenarios

import (
	"fmt"
	"math/rand"
	"time"

	"cache-demo/cache"
	"cache-demo/utils"
)

// KeyGranularityScenario compares three ways to key the same user data.
type KeyGranularityScenario struct{}

func (KeyGranularityScenario) Name() string        { return "Cache key granularity comparison" }
func (KeyGranularityScenario) Description() string { return "Too granular vs too coarse vs balanced keys" }

const users = 10000

// The same user data under every strategy.
var userFields = map[string]string{
	"name": "Alice", "email": "alice@example.com", "avatar": "https://cdn.example.com/a/alice.png",
	"phone": "555-1234", "address": "123 Main St",
}
var fieldOrder = []string{"name", "email", "avatar", "phone", "address"}

// strategy says which cache key holds each field, and which fields share that key.
type strategy struct {
	title, note string
	keyOf       func(uid int, field string) string
}

func pick(m map[string]string, fields ...string) map[string]string {
	out := map[string]string{}
	for _, f := range fields {
		out[f] = m[f]
	}
	return out
}

var strategies = []strategy{
	{"Strategy 1: Too Granular", "one key per field", func(uid int, f string) string { return fmt.Sprintf("user:%d:%s", uid, f) }},
	{"Strategy 2: Too Coarse", "one key per user", func(uid int, f string) string { return fmt.Sprintf("user:%d", uid) }},
	{"Strategy 3: Balanced", "profile + contact keys", func(uid int, f string) string {
		if f == "phone" || f == "address" {
			return fmt.Sprintf("user:%d:contact", uid)
		}
		return fmt.Sprintf("user:%d:profile", uid)
	}},
}

// load simulates a DB read: returns what must be cached under key (all fields sharing it).
func (s strategy) load(uid int, field string) interface{} {
	key := s.keyOf(uid, field)
	var fields []string
	for _, f := range fieldOrder {
		if s.keyOf(uid, f) == key {
			fields = append(fields, f)
		}
	}
	if len(fields) == 1 {
		return userFields[fields[0]] // a single field is stored as a plain string
	}
	return pick(userFields, fields...)
}

// keysFor lists the distinct keys one user occupies under this strategy.
func (s strategy) keysFor(uid int) []string {
	seen := map[string]bool{}
	var keys []string
	for _, f := range fieldOrder {
		if k := s.keyOf(uid, f); !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	return keys
}

// workload replays 1000 operations on ONE user against a real cache:
// 80% reads of a random field, 20% "update email" which invalidates the key holding email.
// Seeded, so output is identical on every run. It is a simulation, not a measurement.
func (s strategy) workload() (hits, misses int64, ratio float64) {
	const uid, ops = 1, 1000
	c := cache.NewCache()
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < ops; i++ {
		if rng.Intn(100) < 20 {
			c.Delete(s.keyOf(uid, "email")) // invalidation on write
			continue
		}
		f := fieldOrder[rng.Intn(len(fieldOrder))]
		key := s.keyOf(uid, f)
		if _, hit := c.Get(key); !hit {
			c.Set(key, s.load(uid, f), time.Hour)
		}
	}
	return c.Stats()
}

func (KeyGranularityScenario) Run() error {
	utils.PrintHeader("Scenario 4: Cache Key Granularity")

	for i, s := range strategies {
		// Memory: really store ONE user under this strategy, measure it with the
		// cache's estimator, then scale to 10,000 users. The total is an ESTIMATE.
		sample := cache.NewCache()
		for _, k := range s.keysFor(5000) {
			var field string
			for _, f := range fieldOrder {
				if s.keyOf(5000, f) == k {
					field = f
					break
				}
			}
			sample.Set(k, s.load(5000, field), time.Hour)
		}
		keysPerUser := sample.Len()
		totalBytes := sample.MemoryUsage() * users

		fmt.Printf("%s (%s)\n", s.title, s.note)
		fmt.Printf("Keys per user: %d  e.g. %v\n", keysPerUser, s.keysFor(123))
		fmt.Printf("For %d users: %d keys\n", users, keysPerUser*users)
		fmt.Printf("Memory: ~%.2f MB (estimate: key + value + %d B/entry overhead)\n",
			float64(totalBytes)/1e6, cache.EntryOverhead)

		hits, misses, ratio := s.workload()
		fmt.Printf("Hit rate: %.0f%% (%d hits, %d misses; simulated: 20%% of ops update email)\n",
			ratio*100, hits, misses)

		switch i {
		case 1:
			utils.PrintWarning("Update email -> the ENTIRE user entry is invalidated (name, avatar, phone, address too)")
		case 2:
			utils.PrintSuccess("Update email -> only user:<id>:profile is invalidated; contact stays cached")
		}
		fmt.Println()
	}

	utils.PrintSeparator()
	fmt.Println("Recommendation: Strategy 3 balances memory, hit rate, and invalidation complexity.")
	fmt.Println("(Sizes and hit rates above come from this demo's model, not from a real Redis.)")
	return nil
}
