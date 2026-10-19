package scenarios

import (
	"fmt"
	"io"
	"time"
)

func Isolated(w io.Writer) error {
	fmt.Fprintln(w, "=== 4. Per-courier isolation and cautious recovery ===")
	fmt.Fprintln(w, "Three failures open a circuit; cooldown is 30 simulated seconds.")
	s := newShop(60 * time.Millisecond)
	s.partners[1].SetDown(true)
	checkout := func(label string) {
		fmt.Fprintln(w, "\n"+label)
		start := time.Now()
		results := s.collect(true)
		elapsed := time.Since(start)
		display(w, results)
		s.totals(w)
		fmt.Fprintf(w, "Measured simulated checkout wait: %v\n", elapsed.Round(time.Millisecond))
	}
	for n := 1; n <= 4; n++ {
		checkout(fmt.Sprintf("Checkout %d", n))
	}
	s.now = s.now.Add(cooldown)
	checkout("Advance clock 30s: one probe to B, still down -> reopen")
	checkout("No clock advance: B is skipped again; A/C still quote")
	s.partners[1].SetDown(false)
	s.now = s.now.Add(cooldown)
	checkout("B recovers; advance clock another 30s: successful probe -> close")
	checkout("Next checkout: all three shipping options restored")
	fmt.Fprintln(w, "One courier going down should remove one option, not take down checkout.")
	fmt.Fprintln(w, "Breakers complement deadlines; they do not cancel in-flight calls or replace concurrency limits.")
	return nil
}
