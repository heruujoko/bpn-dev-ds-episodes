package scenarios

import (
	"fmt"
	"time"

	"rate-limit-demo/limiter"
	"rate-limit-demo/utils"
)

type TokenBucketScenario struct{}

func (TokenBucketScenario) Name() string { return "Token bucket" }
func (TokenBucketScenario) Description() string {
	return "Burst up to five, then refill one token per second"
}
func (TokenBucketScenario) Run() error {
	utils.PrintHeader("Scenario 3: Token Bucket")
	now := time.Unix(0, 0)
	l := limiter.NewTokenBucket(5, 1, func() time.Time { return now })
	fmt.Println("Bucket: capacity 5, refill 1 token/second")
	for i := 1; i <= 6; i++ {
		if ok, retry := l.Allow("bot"); ok {
			utils.PrintSuccess(fmt.Sprintf("t=0.0s request %d: allowed", i))
		} else {
			utils.PrintError(fmt.Sprintf("t=0.0s request %d: 429 (Retry-After: %s)", i, retry))
		}
	}
	now = now.Add(time.Second)
	if ok, _ := l.Allow("bot"); ok {
		utils.PrintSuccess("t=1.0s request 7: allowed (one token refilled)")
	}
	now = now.Add(100 * time.Millisecond)
	if ok, retry := l.Allow("bot"); !ok {
		utils.PrintError(fmt.Sprintf("t=1.1s request 8: 429 (Retry-After: %s)", retry.Round(100*time.Millisecond)))
	}
	utils.PrintSuccess("Unlike fixed windows, a boundary does not restore a full burst.")
	return nil
}
