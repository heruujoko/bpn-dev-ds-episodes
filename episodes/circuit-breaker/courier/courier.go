// Package courier simulates fictional shipping partners; no orders are placed.
package courier

import (
	"context"
	"sync"
)

type Quote struct {
	Name  string
	Price int
}
type Partner struct {
	Name  string
	price int
	mu    sync.Mutex
	down  bool
	calls int
}

func New(name string, price int) *Partner { return &Partner{Name: name, price: price} }
func (p *Partner) SetDown(down bool)      { p.mu.Lock(); defer p.mu.Unlock(); p.down = down }
func (p *Partner) Calls() int             { p.mu.Lock(); defer p.mu.Unlock(); return p.calls }
func (p *Partner) Quote(ctx context.Context) (Quote, error) {
	p.mu.Lock()
	p.calls++
	down := p.down
	p.mu.Unlock()
	if down {
		<-ctx.Done()
		return Quote{}, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return Quote{}, err
	}
	return Quote{Name: p.Name, Price: p.price}, nil
}
