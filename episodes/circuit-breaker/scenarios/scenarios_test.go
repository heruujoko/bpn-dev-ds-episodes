package scenarios

import (
	"circuit-breaker-demo/breaker"
	"errors"
	"testing"
	"time"
)

func counts(s *shop) [3]int {
	return [3]int{s.partners[0].Calls(), s.partners[1].Calls(), s.partners[2].Calls()}
}

func TestCheckoutBoundaries(t *testing.T) {
	for _, mode := range []string{"healthy", "none", "shared", "isolated"} {
		t.Run(mode, func(t *testing.T) {
			s := newShop(time.Millisecond)
			s.partners[1].SetDown(mode != "healthy")
			for i := 0; i < 4; i++ {
				var results []result
				if mode == "shared" {
					var err error
					results, err = s.sharedCheckout()
					if i == 3 {
						if !errors.Is(err, breaker.ErrOpen) || len(results) != 0 {
							t.Fatal("shared did not reject", err)
						}
						continue
					}
				} else {
					results = s.collect(mode == "isolated")
				}
				if len(results) != 3 || results[0].err != nil || results[2].err != nil {
					t.Fatal("lost healthy quotes", results)
				}
				if mode == "healthy" && results[1].err != nil {
					t.Fatal("healthy B failed")
				}
				if mode != "healthy" && results[1].err == nil {
					t.Fatal("down B succeeded")
				}
			}
			want := [3]int{4, 4, 4}
			if mode == "shared" {
				want = [3]int{3, 3, 3}
			}
			if mode == "isolated" {
				want = [3]int{4, 3, 4}
			}
			if got := counts(s); got != want {
				t.Fatalf("calls=%v want=%v", got, want)
			}
		})
	}
}

func TestIsolatedRecovery(t *testing.T) {
	s := newShop(time.Millisecond)
	s.partners[1].SetDown(true)
	for i := 0; i < 4; i++ {
		s.collect(true)
	}
	s.now = s.now.Add(cooldown)
	r := s.collect(true)[1]
	if r.during != breaker.HalfOpen || r.after != breaker.Open || s.partners[1].Calls() != 4 {
		t.Fatal("failed probe did not reopen", r)
	}
	if r = s.collect(true)[1]; !errors.Is(r.err, breaker.ErrOpen) || s.partners[1].Calls() != 4 {
		t.Fatal("cooldown did not reject")
	}
	s.partners[1].SetDown(false)
	s.now = s.now.Add(cooldown)
	r = s.collect(true)[1]
	if r.during != breaker.HalfOpen || r.after != breaker.Closed || r.err != nil {
		t.Fatal("successful probe did not close", r)
	}
	s.collect(true)
	if got := counts(s); got != [3]int{8, 6, 8} {
		t.Fatal("recovery calls", got)
	}
}
