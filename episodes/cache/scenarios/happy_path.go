// Package scenarios contains the four cache demos. Each has Name, Description and Run.
package scenarios

import (
	"fmt"
	"time"

	"cache-demo/cache"
	"cache-demo/utils"
)

// HappyPathScenario: the same slow lookup, with and without a cache.
type HappyPathScenario struct{}

func (HappyPathScenario) Name() string        { return "Cache makes things fast (happy path)" }
func (HappyPathScenario) Description() string { return "5 slow DB reads, with and without a cache" }

const requests = 5

// fetchUserFromDB simulates a slow database query.
func fetchUserFromDB(id int) string {
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("user-%d", id)
}

func (HappyPathScenario) Run() error {
	utils.PrintHeader("Scenario 1: Cache Makes Things Fast")

	// Without cache: every request pays the full DB price.
	fmt.Println("Without cache:")
	var totalNoCache time.Duration
	for i := 1; i <= requests; i++ {
		start := time.Now()
		fetchUserFromDB(123)
		d := time.Since(start)
		totalNoCache += d
		utils.PrintError(fmt.Sprintf("Request %d: %v (cache miss)", i, d.Round(time.Millisecond)))
	}
	fmt.Printf("Total: %v\n", totalNoCache.Round(time.Millisecond))
	fmt.Println()
	utils.PrintSeparator()

	// With cache: only the first request hits the DB; the rest are map lookups.
	fmt.Println("With cache:")
	c := cache.NewCache()
	var totalCache time.Duration
	for i := 1; i <= requests; i++ {
		start := time.Now()
		user, hit := c.Get("user:123")
		if !hit {
			user = fetchUserFromDB(123)
			c.Set("user:123", user, time.Minute)
		}
		_ = user
		d := time.Since(start)
		totalCache += d
		if hit {
			utils.PrintSuccess(fmt.Sprintf("Request %d: %.3fms (cache hit)", i, float64(d)/float64(time.Millisecond)))
		} else {
			utils.PrintError(fmt.Sprintf("Request %d: %v (cache miss, stored)", i, d.Round(time.Millisecond)))
		}
	}
	fmt.Printf("Total: %v\n", totalCache.Round(time.Millisecond)) // dominated by the single miss
	fmt.Println()
	utils.PrintSeparator()

	hits, misses, _ := c.Stats()
	utils.PrintStats(hits, misses)
	fmt.Printf("Time: %v without cache vs %v with cache (%.1fx faster)\n",
		totalNoCache.Round(time.Millisecond), totalCache.Round(time.Millisecond),
		float64(totalNoCache)/float64(totalCache))
	return nil
}
