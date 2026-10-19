package scenarios

import (
	"fmt"
	"io"
	"time"
)

func Shared(w io.Writer) error {
	fmt.Fprintln(w, "=== 3. Wrong boundary: one breaker around the whole checkout ===")
	fmt.Fprintln(w, "Deliberately wrong policy: ANY missing quote makes the aggregate operation fail.")
	s := newShop(60 * time.Millisecond)
	s.partners[1].SetDown(true)
	for n := 1; n <= 4; n++ {
		fmt.Fprintf(w, "\nCheckout %d\n", n)
		before := s.shared.State()
		results, err := s.sharedCheckout()
		display(w, results)
		if err != nil {
			fmt.Fprintln(w, "Aggregate outcome:", err)
		}
		fmt.Fprintf(w, "Shared circuit: %s -> %s\n", before, s.shared.State())
	}
	s.totals(w)
	fmt.Fprintln(w, "B's failures blocked healthy A and C too. The operation boundary is the mistake.")
	return nil
}
