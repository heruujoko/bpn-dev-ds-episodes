package scenarios

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestScaleOutLocalCounters(t *testing.T) {
	now := time.Unix(0, 0)
	for _, test := range []struct {
		name          string
		route         func(int) int
		local, divide int
	}{
		{name: "round-robin", route: scaleOutRoutings[0].route, local: 15, divide: 3},
		{name: "weighted", route: scaleOutRoutings[1].route, local: 11, divide: 3},
		{name: "sticky", route: scaleOutRoutings[2].route, local: 5, divide: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := localAllowed(scaleOutLimit, scaleOutRequests, test.route, func() time.Time { return now }); got != test.local {
				t.Fatalf("local allowed = %d, want %d", got, test.local)
			}
			if got := localAllowed(scaleOutLimit/scaleOutServers, scaleOutRequests, test.route, func() time.Time { return now }); got != test.divide {
				t.Fatalf("divided allowed = %d, want %d", got, test.divide)
			}
		})
	}
}

func TestScaleOutRedisCounter(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is not set")
	}
	if err := redisReachable(addr); err != nil {
		t.Skipf("Redis is unreachable at %s: %v", addr, err)
	}

	for _, routing := range scaleOutRoutings {
		t.Run(routing.name, func(t *testing.T) {
			key := fmt.Sprintf("ratelimit:test:scaleout:%d:%s", time.Now().UnixNano(), routing.name)
			got, err := redisAllowed(context.Background(), addr, key, scaleOutRequests, routing.route)
			if err != nil {
				t.Fatal(err)
			}
			if got != scaleOutLimit {
				t.Fatalf("Redis allowed = %d, want %d", got, scaleOutLimit)
			}
		})
	}
}
