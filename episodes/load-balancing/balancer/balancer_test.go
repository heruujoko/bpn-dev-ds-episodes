package balancer

import "testing"

func TestRouting(t *testing.T) {
	for _, tc := range []struct {
		name   string
		choose func(int) (int, error)
		cycle  []int
	}{
		{"round-robin", func(i int) (int, error) { return RoundRobin(3, i) }, []int{0, 1, 2}},
		{"weighted", func(i int) (int, error) { return Weighted([]int{3, 2, 1}, i) }, []int{0, 0, 0, 1, 1, 2}},
		{"zero weight", func(i int) (int, error) { return Weighted([]int{0, 2, 0, 1}, i) }, []int{1, 1, 3}},
		{"one backend", func(i int) (int, error) { return RoundRobin(1, i) }, []int{0}},
		{"one weighted backend", func(i int) (int, error) { return Weighted([]int{4}, i) }, []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 24; i++ {
				got, err := tc.choose(i)
				if err != nil || got != tc.cycle[i%len(tc.cycle)] {
					t.Fatalf("request %d: backend=%d err=%v", i, got, err)
				}
			}
		})
	}
}

func TestTotals(t *testing.T) {
	var rr, weighted [3]int
	for i := 0; i < 12; i++ {
		a, err := RoundRobin(3, i)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Weighted([]int{3, 2, 1}, i)
		if err != nil {
			t.Fatal(err)
		}
		rr[a]++
		weighted[b]++
	}
	if rr != [3]int{4, 4, 4} || weighted != [3]int{6, 4, 2} {
		t.Fatalf("totals: round-robin=%v weighted=%v", rr, weighted)
	}
}

func TestInvalidInputs(t *testing.T) {
	for _, n := range []int{0, -1} {
		if _, err := RoundRobin(n, 0); err == nil {
			t.Fatal("expected backend count error")
		}
	}
	if _, err := RoundRobin(3, -1); err == nil {
		t.Fatal("expected index error")
	}
	maxInt := int(^uint(0) >> 1)
	for _, weights := range [][]int{nil, {}, {0, 0}, {-1, 2}, {maxInt, 1}} {
		if _, err := Weighted(weights, 0); err == nil {
			t.Fatalf("accepted weights %v", weights)
		}
	}
	if _, err := Weighted([]int{1}, -1); err == nil {
		t.Fatal("expected index error")
	}
	if got, err := Weighted([]int{maxInt}, maxInt); err != nil || got != 0 {
		t.Fatalf("valid maximum weight: got %d, err %v", got, err)
	}
}
