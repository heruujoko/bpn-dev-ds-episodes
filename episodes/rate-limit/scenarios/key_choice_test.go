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

	balanced := limiter.NewFixedWindow(2, time.Minute, clock)
	for range 2 {
		if !allowLoginAttempt(balanced, "alice@example.com", "203.0.113.7") {
			t.Fatal("first attempts should be allowed")
		}
	}
	if allowLoginAttempt(balanced, "alice@example.com", "203.0.113.7") {
		t.Fatal("account plus IP key should enforce its limit")
	}
}
