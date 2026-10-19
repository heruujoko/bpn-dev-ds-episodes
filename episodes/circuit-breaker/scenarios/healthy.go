package scenarios

import (
	"fmt"
	"io"
	"time"
)

func Healthy(w io.Writer) error {
	fmt.Fprintln(w, "=== 1. Healthy checkout: three courier options ===")
	fmt.Fprintln(w, "Local simulation, concurrent calls, individual 60ms deadlines. Prices are fictional.")
	s := newShop(60 * time.Millisecond)
	for n := 1; n <= 4; n++ {
		fmt.Fprintf(w, "\nCheckout %d\n", n)
		display(w, s.collect(false))
	}
	s.totals(w)
	return nil
}
