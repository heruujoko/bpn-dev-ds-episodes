package scenarios

import (
	"fmt"
	"time"

	"rate-limit-demo/limiter"
	"rate-limit-demo/utils"
)

type FixedWindowScenario struct{}

func (FixedWindowScenario) Name() string { return "Fixed window (and its boundary burst)" }
func (FixedWindowScenario) Description() string {
	return "Five per second per client; ten can pass around a window edge"
}
func (FixedWindowScenario) Run() error {
	utils.PrintHeader("Scenario 2: Fixed Window")
	now := time.Unix(0, 900*time.Millisecond.Nanoseconds())
	l := limiter.NewFixedWindow(5, time.Second, func() time.Time { return now })
	fmt.Println("Limit: 5 requests per 1s window, per client")
	for i := 1; i <= 6; i++ {
		if ok, retry := l.Allow("bot"); ok {
			utils.PrintSuccess(fmt.Sprintf("Bot request %d: allowed", i))
		} else {
			utils.PrintError(fmt.Sprintf("Bot request %d: 429 Too Many Requests (Retry-After: %s)", i, retry))
		}
	}
	for i := 1; i <= 3; i++ {
		ok, _ := l.Allow("alice")
		if ok {
			utils.PrintSuccess(fmt.Sprintf("Alice request %d: allowed, fast", i))
		}
	}
	utils.PrintSeparator()
	fmt.Println("Boundary burst: five at t=0.9s, then five at t=1.1s.")
	now = now.Add(200 * time.Millisecond)
	allowed := 0
	for range 5 {
		if ok, _ := l.Allow("bot"); ok {
			allowed++
		}
	}
	utils.PrintWarning(fmt.Sprintf("%d extra bot requests allowed after the window reset: 10 in 0.2s.", allowed))
	return nil
}
