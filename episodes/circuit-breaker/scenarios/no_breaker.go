package scenarios

import (
	"fmt"
	"io"
	"time"
)

func NoBreaker(w io.Writer) error {
	fmt.Fprintln(w, "=== 2. B is down: no circuit breaker ===")
	fmt.Fprintln(w, "Each checkout waits for all quotes or their individual deadlines.")
	s := newShop(60 * time.Millisecond)
	s.partners[1].SetDown(true)
	for n := 1; n <= 4; n++ {
		fmt.Fprintf(w, "\nCheckout %d\n", n)
		start := time.Now()
		results := s.collect(false)
		display(w, results)
		fmt.Fprintf(w, "Measured simulated checkout wait: %v\n", time.Since(start).Round(time.Millisecond))
	}
	s.totals(w)
	fmt.Fprintln(w, "Timeout bounds each wait. It does not remember B's previous failures.")
	return nil
}
