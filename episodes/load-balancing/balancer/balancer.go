// Package balancer selects simulated backends, not network destinations.
package balancer

import "fmt"

func RoundRobin(backends, requestIndex int) (int, error) {
	if backends <= 0 || requestIndex < 0 {
		return 0, fmt.Errorf("backend count must be positive and request index nonnegative")
	}
	return requestIndex % backends, nil
}

// Weighted uses grouped weighted round-robin: [3,2,1] repeats A,A,A,B,B,C.
// ponytail: static weights only; add a separate adaptive example for live load.
func Weighted(weights []int, requestIndex int) (int, error) {
	if requestIndex < 0 {
		return 0, fmt.Errorf("request index must be nonnegative")
	}
	total := 0
	maxInt := int(^uint(0) >> 1)
	for _, w := range weights {
		if w < 0 || w > maxInt-total {
			return 0, fmt.Errorf("weights must be nonnegative with a sum that fits an int")
		}
		total += w
	}
	if total == 0 {
		return 0, fmt.Errorf("at least one weight must be positive")
	}
	position := requestIndex % total
	for backend, w := range weights {
		if position < w {
			return backend, nil
		}
		position -= w
	}
	return 0, fmt.Errorf("no backend selected")
}
