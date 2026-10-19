// Package breaker is a teaching state machine, not a production resilience library.
package breaker

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type State string

const (
	Closed   State = "closed"
	Open     State = "open"
	HalfOpen State = "half-open"
)

var ErrOpen = errors.New("circuit open: dependency call skipped")

type Breaker struct {
	mu                  sync.Mutex
	state               State
	failures, threshold int
	cooldown            time.Duration
	until               time.Time
	now                 func() time.Time
	generation          uint64
}

func New(threshold int, cooldown time.Duration, now func() time.Time) (*Breaker, error) {
	if threshold <= 0 || cooldown <= 0 || now == nil {
		return nil, fmt.Errorf("positive threshold/cooldown and a clock are required")
	}
	return &Breaker{state: Closed, threshold: threshold, cooldown: cooldown, now: now}, nil
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Execute counts returned errors as failures; callers must choose an appropriate
// operation boundary and error classification. No lock is held during the call.
// ponytail: callbacks must return normally; use a production library for panic handling.
func (b *Breaker) Execute(call func() error) error {
	b.mu.Lock()
	if b.state == Open {
		if b.now().Before(b.until) {
			b.mu.Unlock()
			return ErrOpen
		}
		b.state = HalfOpen
		b.generation++
	} else if b.state == HalfOpen {
		b.mu.Unlock()
		return ErrOpen
	}
	generation := b.generation
	probe := b.state == HalfOpen
	b.mu.Unlock()

	err := call()
	b.mu.Lock()
	defer b.mu.Unlock()
	if generation != b.generation {
		return err
	} // older in-flight result cannot undo a transition
	if probe {
		if err == nil {
			b.state = Closed
			b.failures = 0
			b.generation++
		} else {
			b.open()
		}
	} else if err == nil {
		b.failures = 0
	} else {
		b.failures++
		if b.failures >= b.threshold {
			b.open()
		}
	}
	return err
}

// open runs under the mutex and starts a new generation/cooldown.
func (b *Breaker) open() {
	b.state = Open
	b.until = b.now().Add(b.cooldown)
	b.generation++
}
