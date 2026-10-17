package scenarios

import (
	"testing"
	"time"
)

func TestPerServerCountersAllowFifteen(t *testing.T) {
	now := time.Unix(0, 0)
	if got := allowedAcrossServers(5, 15, func() time.Time { return now }); got != 15 {
		t.Fatalf("allowed = %d, want 15", got)
	}
}
