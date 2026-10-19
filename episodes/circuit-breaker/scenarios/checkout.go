package scenarios

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"circuit-breaker-demo/breaker"
	"circuit-breaker-demo/courier"
)

const cooldown = 30 * time.Second

type result struct {
	name                  string
	quote                 courier.Quote
	err                   error
	before, during, after breaker.State
}

type shop struct {
	partners []*courier.Partner
	circuits []*breaker.Breaker
	shared   *breaker.Breaker
	now      time.Time
	timeout  time.Duration
}

func newShop(timeout time.Duration) *shop {
	s := &shop{partners: []*courier.Partner{courier.New("A", 12000), courier.New("B", 10000), courier.New("C", 15000)}, now: time.Unix(0, 0), timeout: timeout}
	// The fake clock advances only between completed checkouts.
	clock := func() time.Time { return s.now }
	s.shared, _ = breaker.New(3, cooldown, clock)
	for range s.partners {
		b, _ := breaker.New(3, cooldown, clock)
		s.circuits = append(s.circuits, b)
	}
	return s
}

func (s *shop) collect(isolated bool) []result {
	results := make([]result, len(s.partners))
	var wg sync.WaitGroup
	for i, p := range s.partners {
		wg.Add(1)
		go func(i int, p *courier.Partner) {
			defer wg.Done()
			r := result{name: p.Name}
			call := func() error {
				ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
				defer cancel()
				var err error
				r.quote, err = p.Quote(ctx)
				return err
			}
			if isolated {
				b := s.circuits[i]
				r.before = b.State()
				r.err = b.Execute(func() error { r.during = b.State(); return call() })
				r.after = b.State()
			} else {
				r.err = call()
			}
			results[i] = r
		}(i, p)
	}
	wg.Wait()
	return results
}

func (s *shop) sharedCheckout() ([]result, error) {
	var results []result
	err := s.shared.Execute(func() error {
		results = s.collect(false)
		for _, r := range results {
			if r.err != nil {
				return fmt.Errorf("incomplete checkout: courier %s: %w", r.name, r.err)
			}
		}
		return nil
	})
	return results, err
}

func display(w io.Writer, results []result) {
	available := 0
	for _, r := range results {
		if r.err == nil {
			fmt.Fprintf(w, "Courier %s: quote IDR %d\n", r.name, r.quote.Price)
			available++
		} else {
			fmt.Fprintf(w, "Courier %s: temporarily unavailable (%v)\n", r.name, r.err)
		}
		if r.before != "" {
			if r.during != "" && r.during != r.before {
				fmt.Fprintf(w, "  circuit: %s -> %s -> %s\n", r.before, r.during, r.after)
			} else {
				fmt.Fprintf(w, "  circuit: %s -> %s\n", r.before, r.after)
			}
		}
	}
	if available == 0 {
		fmt.Fprintln(w, "No shipping option currently available; cart saved. No booking made.")
	} else {
		fmt.Fprintf(w, "%d shipping options available. No automatic courier selection or booking.\n", available)
	}
}

func (s *shop) totals(w io.Writer) {
	fmt.Fprintf(w, "Dependency calls: A=%d B=%d C=%d\n", s.partners[0].Calls(), s.partners[1].Calls(), s.partners[2].Calls())
}
