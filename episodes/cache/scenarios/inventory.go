package scenarios

import (
	"fmt"
	"sync"
	"time"

	"cache-demo/cache"
	"cache-demo/utils"
)

// InventoryScenario: two buyers trust a cached stock count and oversell.
type InventoryScenario struct{}

func (InventoryScenario) Name() string        { return "Anti-pattern: Caching inventory" }
func (InventoryScenario) Description() string { return "Two buyers, one item, stale cached stock" }

func (InventoryScenario) Run() error {
	utils.PrintHeader("Scenario 3: Anti-Pattern - Caching Inventory")

	var (
		dbMu    sync.Mutex
		dbStock = 1 // the real stock lives in the "database"
	)
	c := cache.NewCache()
	c.Set("product:456:stock", dbStock, 10*time.Second)

	fmt.Println("Initial state:")
	fmt.Printf("Stock: %d\n", dbStock)
	fmt.Println("Cache: 1 (TTL: 10s)")
	fmt.Println()

	// The race is made deterministic: both buyers must READ the cache before
	// EITHER writes the DB. That is exactly the interleaving a busy site hits
	// regularly, just without relying on scheduler luck.
	var allRead, done sync.WaitGroup
	allRead.Add(2)
	done.Add(2)

	results := make([]string, 2)
	buyer := func(idx int, name string) {
		defer done.Done()
		v, _ := c.Get("product:456:stock") // 1. check the (stale) cache
		seen := v.(int)
		allRead.Done() // 2. announce "I have read"...
		allRead.Wait() //    ...and wait until the other buyer has read too

		if seen > 0 { // 3. both see 1, so both buy
			dbMu.Lock()
			dbStock-- // the DB is updated, the cache is not
			dbMu.Unlock()
			results[idx] = fmt.Sprintf("Request %s: checks stock -> cache says %d -> buys", name, seen)
		}
	}
	go buyer(0, "A")
	go buyer(1, "B")
	done.Wait()

	for _, r := range results {
		utils.PrintSuccess(r)
	}
	fmt.Println()

	fmt.Println("Result:")
	fmt.Printf("Stock: %d\n", dbStock)
	if dbStock < 0 {
		utils.PrintCritical("OVERSOLD! Both requests succeeded because cache showed stale data")
	}
	fmt.Println()
	fmt.Println("Lesson: Don't cache frequently-changing data with business consequences.")
	return nil
}
