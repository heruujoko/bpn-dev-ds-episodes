package scenarios

import (
	"fmt"
	"time"

	"cache-demo/cache"
	"cache-demo/utils"
)

// PermissionsScenario: a long TTL keeps a revoked permission alive.
type PermissionsScenario struct{}

func (PermissionsScenario) Name() string        { return "Anti-pattern: Caching permissions" }
func (PermissionsScenario) Description() string { return "A revoked permission survives in the cache" }

func (PermissionsScenario) Run() error {
	utils.PrintHeader("Scenario 2: Anti-Pattern - Caching Permissions")

	c := cache.NewCache()
	db := []string{"read", "write"} // our fake "database"

	// Step 1: login loads permissions from the DB and caches them for an hour.
	fmt.Println("Step 1: User logs in")
	c.Set("user:123:perms", append([]string(nil), db...), time.Hour)
	fmt.Printf("Permissions cached: %v\n", db)
	fmt.Println("TTL: 1 hour")
	fmt.Println()

	// Step 2: admin revokes "write" in the DB. Nobody touches the cache.
	fmt.Println("Step 2: Admin revokes 'write' permission")
	db = []string{"read"}
	fmt.Printf("Database updated: %v\n", db)
	utils.PrintWarning("Cache NOT invalidated!")
	fmt.Println()

	// Step 3: the next request reads the cache and sees the old permissions.
	fmt.Println("Step 3: User makes request")
	v, _ := c.Get("user:123:perms")
	perms := v.([]string)
	utils.PrintWarning(fmt.Sprintf("Cache returns: %v  (DB says %v) STALE!", perms, db))
	utils.PrintCritical("Security breach! User has 'write' access they shouldn't have!")
	fmt.Println()

	fmt.Println("Lesson: Don't cache security-critical data with long TTL.")
	return nil
}
