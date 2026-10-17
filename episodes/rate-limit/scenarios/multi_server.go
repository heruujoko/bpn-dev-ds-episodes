package scenarios

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"rate-limit-demo/limiter"
	"rate-limit-demo/utils"
)

type MultiServerScenario struct{}

func (MultiServerScenario) Name() string { return "Multiple servers: in-memory vs Redis" }
func (MultiServerScenario) Description() string {
	return "Three servers multiply an in-memory limit; Redis shares it"
}

const redisOpTimeout = time.Second

func allowedAcrossServers(limit, requests int, now func() time.Time) int {
	servers := []*limiter.FixedWindow{
		limiter.NewFixedWindow(limit, time.Minute, now),
		limiter.NewFixedWindow(limit, time.Minute, now),
		limiter.NewFixedWindow(limit, time.Minute, now),
	}
	allowed := 0
	for i := range requests {
		if ok, _ := servers[i%len(servers)].Allow("bot"); ok {
			allowed++
		}
	}
	return allowed
}

var incrementWithExpiry = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

func redisReachable(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	probe := redis.NewClient(&redis.Options{Addr: addr})
	defer probe.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return probe.Ping(ctx).Err()
}

func (MultiServerScenario) Run() error {
	utils.PrintHeader("Scenario 5: Multiple Servers - In-Memory vs Redis")
	const (
		limit    = 5
		requests = 15
		window   = time.Minute
	)
	now := time.Unix(0, 0)

	fmt.Printf("Part 1: %d servers, each with its own in-memory limit of %d per window\n", 3, limit)
	allowed := allowedAcrossServers(limit, requests, func() time.Time { return now })
	utils.PrintError(fmt.Sprintf("%d/%d bot requests got through. Limit was %d.", allowed, requests, limit))
	utils.PrintSeparator()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	if err := redisReachable(addr); err != nil {
		utils.PrintWarning(fmt.Sprintf("Redis is not reachable at %s (%v)", addr, err))
		fmt.Println("This scenario needs Redis. Start it with: docker compose run --rm demo")
		return nil
	}

	fmt.Printf("Part 2: 3 servers share one Redis counter with a limit of %d per window\n", limit)
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()
	prefix := fmt.Sprintf("ratelimit:demo:%d", time.Now().UnixNano())
	key := prefix + ":bot"
	ctx := context.Background()
	defer client.Del(ctx, key)
	allowed = 0
	for range requests {
		requestCtx, cancel := context.WithTimeout(ctx, redisOpTimeout)
		// INCR and first-key expiry must be one operation: a crash between separate
		// commands would leave a key that never expires and limit the user forever.
		count, err := incrementWithExpiry.Run(requestCtx, client, []string{key}, window.Milliseconds()).Int()
		cancel()
		if err != nil {
			return fmt.Errorf("increment shared counter: %w", err)
		}
		if count <= limit {
			allowed++
		}
	}
	utils.PrintSuccess(fmt.Sprintf("%d/%d bot requests got through. Shared limit was %d.", allowed, requests, limit))
	fmt.Println()
	fmt.Println("Redis trade-offs:")
	fmt.Println("  - a network hop per request")
	fmt.Println("  - fail-open vs fail-closed is a business decision")
	fmt.Println("  - counter expiry must be atomic")
	return nil
}
