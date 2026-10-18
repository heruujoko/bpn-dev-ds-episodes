package scenarios

import (
	"fmt"
	"load-balancing-demo/balancer"
)

func Weighted() error {
	fmt.Println("=== Scenario 2: Different Capacities -> Weighted Routing ===")
	fmt.Println("Simulation: configured weights A=3 B=2 C=1, not measured live capacity.")
	if err := printAssignments("Round-robin baseline", func(i int) (int, error) { return balancer.RoundRobin(3, i) }); err != nil {
		return err
	}
	if err := printAssignments("Grouped weighted round-robin: A, A, A, B, B, C, repeat", func(i int) (int, error) { return balancer.Weighted([]int{3, 2, 1}, i) }); err != nil {
		return err
	}
	fmt.Println("Grouped routing can create bursts; this is not smooth weighted round-robin.")
	fmt.Println("Weights need workload evidence. Neither policy detects slow or failed servers.")
	fmt.Println("Callback: routing changes do not make local caches or rate-limit counters shared.")
	return nil
}
