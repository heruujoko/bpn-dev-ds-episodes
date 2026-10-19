package breaker

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

var failure = errors.New("dependency unavailable")

func TestTransitions(t *testing.T) {
	now := time.Unix(0, 0)
	b, err := New(3, time.Second, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	fail := func() error { return failure }
	ok := func() error { return nil }
	b.Execute(fail)
	b.Execute(fail)
	b.Execute(ok)
	b.Execute(fail)
	b.Execute(fail)
	if b.State() != Closed {
		t.Fatal("success must reset consecutive failures")
	}
	b.Execute(fail)
	if b.State() != Open {
		t.Fatal("threshold did not open")
	}
	called := false
	check := func() error { called = true; return nil }
	now = now.Add(time.Second - time.Nanosecond)
	if !errors.Is(b.Execute(check), ErrOpen) || called {
		t.Fatal("open called dependency")
	}
	now = now.Add(time.Nanosecond)
	if b.State() != Open {
		t.Fatal("inspection must not reserve a probe")
	}
	if !errors.Is(b.Execute(func() error {
		if b.State() != HalfOpen {
			t.Fatal("probe not half-open")
		}
		return failure
	}), failure) || b.State() != Open {
		t.Fatal("failed probe must reopen")
	}
	if !errors.Is(b.Execute(ok), ErrOpen) {
		t.Fatal("failed probe must restart cooldown")
	}
	now = now.Add(time.Second)
	if err := b.Execute(ok); err != nil || b.State() != Closed {
		t.Fatal("successful probe must close", err)
	}
	if err := b.Execute(ok); err != nil {
		t.Fatal(err)
	}
}

func TestOnlyOneProbe(t *testing.T) {
	var ticks atomic.Int64
	b, _ := New(1, time.Second, func() time.Time { return time.Unix(0, ticks.Load()) })
	b.Execute(func() error { return failure })
	ticks.Store(int64(time.Second))
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() { done <- b.Execute(func() error { close(entered); <-release; return nil }) }()
	<-entered
	err := b.Execute(func() error { return errors.New("must not run") })
	close(release)
	probeErr := <-done
	if !errors.Is(err, ErrOpen) || probeErr != nil || b.State() != Closed {
		t.Fatal("probe gate failed", err, probeErr)
	}
}

func TestOldCompletionsIgnored(t *testing.T) {
	for _, lateErr := range []error{nil, failure} {
		b, _ := New(1, time.Second, time.Now)
		entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		go func() { done <- b.Execute(func() error { close(entered); <-release; return lateErr }) }()
		<-entered
		b.Execute(func() error { return failure })
		close(release)
		<-done
		if b.State() != Open {
			t.Fatal("old completion changed open generation")
		}
	}
}

func TestInvalidConfig(t *testing.T) {
	for _, threshold := range []int{0, -1} {
		if _, err := New(threshold, time.Second, time.Now); err == nil {
			t.Fatal("accepted invalid threshold")
		}
	}
	for _, cooldown := range []time.Duration{0, -1} {
		if _, err := New(1, cooldown, time.Now); err == nil {
			t.Fatal("accepted invalid cooldown")
		}
	}
	if _, err := New(1, time.Second, nil); err == nil {
		t.Fatal("accepted nil clock")
	}
}
