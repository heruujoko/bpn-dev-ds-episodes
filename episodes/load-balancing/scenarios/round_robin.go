package scenarios

import (
	"fmt"
	"load-balancing-demo/balancer"
)

// printAssignments computes totals from the same routing decisions it displays.
func printAssignments(title string, choose func(int) (int, error)) error {
	fmt.Println(title)
	var totals [3]int
	for i := 0; i < 12; i++ {
		backend, err := choose(i)
		if err != nil {
			return err
		}
		if backend < 0 || backend >= len(totals) {
			return fmt.Errorf("invalid backend %d", backend)
		}
		totals[backend]++
		fmt.Printf("Request %2d -> %c\n", i+1, 'A'+backend)
	}
	fmt.Printf("Totals: A=%d B=%d C=%d\n\n", totals[0], totals[1], totals[2])
	return nil
}

func RoundRobin() error {
	fmt.Println("=== Scenario 1: One Server -> Round-Robin ===")
	fmt.Println("Simulation: three backends in one process; no HTTP or timing measurements.")
	if err := printAssignments("Without balancing: all requests to A", func(int) (int, error) { return 0, nil }); err != nil {
		return err
	}
	if err := printAssignments("Round-robin: A, B, C, repeat", func(i int) (int, error) { return balancer.RoundRobin(3, i) }); err != nil {
		return err
	}
	fmt.Println("Equal request counts do not mean equal work: request durations may differ.")
	return nil
}
