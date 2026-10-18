package scenarios

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"rate-limit-demo/limiter"
	"rate-limit-demo/utils"
)

type ScaleOutScenario struct{}

func (ScaleOutScenario) Name() string { return "Scaling out: local counters vs centralized Redis" }
func (ScaleOutScenario) Description() string {
	return "Load-balancer routing changes local limits; Redis does not"
}

const (
	scaleOutServers  = 3
	scaleOutLimit    = 5
	scaleOutRequests = 15
)

type scaleOutRouting struct {
	name  string
	route func(int) int
}

var scaleOutRoutings = []scaleOutRouting{
	{name: "round-robin", route: func(i int) int { return i % scaleOutServers }},
	{name: "weighted 3:2:1", route: func(i int) int { return [...]int{0, 0, 0, 1, 1, 2}[i%6] }},
	{name: "sticky", route: func(int) int { return 0 }},
}

// countAllowed sends requests through route and returns how many allow accepts.
func countAllowed(requests int, route func(int) int, allow func(server int) bool) int {
	allowed := 0
	for i := range requests {
		if allow(route(i)) {
			allowed++
		}
	}
	return allowed
}

func localAllowed(limit, requests int, route func(int) int, now func() time.Time) int {
	servers := make([]*limiter.FixedWindow, scaleOutServers)
	for i := range servers {
		servers[i] = limiter.NewFixedWindow(limit, time.Minute, now)
	}
	return countAllowed(requests, route, func(server int) bool {
		ok, _ := servers[server].Allow("bot")
		return ok
	})
}

func redisAllowed(ctx context.Context, addr, key string, requests int, route func(int) int) (int, error) {
	clients := make([]*redis.Client, scaleOutServers)
	for i := range clients {
		clients[i] = redis.NewClient(&redis.Options{Addr: addr})
		defer clients[i].Close()
	}
	defer clients[0].Del(ctx, key)

	var runErr error
	allowed := countAllowed(requests, route, func(server int) bool {
		if runErr != nil {
			return false
		}
		requestCtx, cancel := context.WithTimeout(ctx, redisOpTimeout)
		count, err := incrementWithExpiry.Run(requestCtx, clients[server], []string{key}, time.Minute.Milliseconds()).Int()
		cancel()
		if err != nil {
			runErr = fmt.Errorf("increment shared counter: %w", err)
			return false
		}
		return count <= scaleOutLimit
	})
	return allowed, runErr
}

func scaleOutCell(allowed int) string {
	mark := "✗"
	if allowed == scaleOutLimit {
		mark = "✓"
	}
	return fmt.Sprintf("%d/%d %s", allowed, scaleOutRequests, mark)
}

func (ScaleOutScenario) Run() error {
	utils.PrintHeader("Scenario 6: Scaling Out - Local Counters vs Centralized Redis")
	now := time.Unix(0, 0)
	localResults := make([]int, len(scaleOutRoutings))
	dividedResults := make([]int, len(scaleOutRoutings))
	perServerLimit := scaleOutLimit / scaleOutServers
	for i, routing := range scaleOutRoutings {
		localResults[i] = localAllowed(scaleOutLimit, scaleOutRequests, routing.route, func() time.Time { return now })
		dividedResults[i] = localAllowed(perServerLimit, scaleOutRequests, routing.route, func() time.Time { return now })
	}

	fmt.Printf("%-30s %-17s %-17s %-17s\n", "strategy", "round-robin", "weighted 3:2:1", "sticky")
	fmt.Println("-----------------------------------------------------------------------------------")
	fmt.Printf("%-30s %-17s %-17s %-17s\n", "Local, 5 per server", scaleOutCell(localResults[0]), scaleOutCell(localResults[1]), scaleOutCell(localResults[2]))
	fmt.Printf("%-30s %-17s %-17s %-17s\n", fmt.Sprintf("Local, 5÷3 per server (%d)", perServerLimit), scaleOutCell(dividedResults[0]), scaleOutCell(dividedResults[1]), scaleOutCell(dividedResults[2]))

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	if err := redisReachable(addr); err != nil {
		fmt.Printf("%-30s needs Redis: docker compose run --rm demo\n", "Redis, 5 shared")
		printScaleOutTakeaways()
		return nil
	}

	redisResults := make([]int, len(scaleOutRoutings))
	prefix := fmt.Sprintf("ratelimit:demo:scaleout:%d", time.Now().UnixNano())
	for i, routing := range scaleOutRoutings {
		allowed, err := redisAllowed(context.Background(), addr, prefix+":"+routing.name, scaleOutRequests, routing.route)
		if err != nil {
			return err
		}
		redisResults[i] = allowed
	}
	fmt.Printf("%-30s %-17s %-17s %-17s\n", "Redis, 5 shared", scaleOutCell(redisResults[0]), scaleOutCell(redisResults[1]), scaleOutCell(redisResults[2]))
	printScaleOutTakeaways()
	return nil
}

func printScaleOutTakeaways() {
	utils.PrintWarning("Local counters make the limit depend on routing and server count; autoscaling to 6 doubles it.")
	utils.PrintWarning("Dividing by N rounds badly, blocks sticky users, and goes stale on every scale event.")
	utils.PrintSuccess("The centralized counter alone is routing-independent, at the cost of a network hop and a dependency.")
}
