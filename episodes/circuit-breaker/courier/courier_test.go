package courier

import (
	"context"
	"errors"
	"testing"
)

func TestQuoteAndCancellation(t *testing.T) {
	c := New("A", 12000)
	q, err := c.Quote(context.Background())
	if err != nil || q.Price != 12000 || q.Name != "A" || c.Calls() != 1 {
		t.Fatal(q, err)
	}
	c.SetDown(true)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := c.Quote(ctx); done <- err }()
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	c.SetDown(false)
	if _, err := c.Quote(context.Background()); err != nil || c.Calls() != 3 {
		t.Fatal("recovery failed", err)
	}
}
