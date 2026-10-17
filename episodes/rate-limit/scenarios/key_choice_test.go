package scenarios

import (
	"testing"
	"time"

	"rate-limit-demo/limiter"
)

func TestLoginKeyChoice(t *testing.T) {
	now := time.Unix(0, 0)
	clock := func() time.Time { return now }

	byUserID := limiter.NewFixedWindow(2, time.Minute, clock)
	for range 100 {
		if !allowByUserID(byUserID, "") {
			t.Fatal("empty user ID should bypass the broken user-ID limiter")
		}
	}

	// Rotating IPs must not reset the per-account budget.
	balanced := limiter.NewFixedWindow(2, time.Minute, clock)
	for _, ip := range []string{"198.51.100.1", "198.51.100.2"} {
		if !allowLoginAttempt(balanced, "alice@example.com", ip) {
			t.Fatal("first attempts should be allowed")
		}
	}
	if allowLoginAttempt(balanced, "alice@example.com", "198.51.100.3") {
		t.Fatal("per-account limit should block a new IP aimed at the same account")
	}

	// One IP spraying many accounts must hit the per-IP limit.
	spray := limiter.NewFixedWindow(2, time.Minute, clock)
	allowLoginAttempt(spray, "a@example.com", "203.0.113.7")
	allowLoginAttempt(spray, "b@example.com", "203.0.113.7")
	if allowLoginAttempt(spray, "c@example.com", "203.0.113.7") {
		t.Fatal("per-IP limit should block one IP trying many accounts")
	}
}
