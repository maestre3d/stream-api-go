package persistence

import (
	"context"
	"errors"
	"sync"

	"github.com/maestre3d/stream-api/internal/aggregate"
)

type Order interface {
	Save(context.Context, aggregate.Order) error
	Get(context.Context, string) (aggregate.Order, error)
}

type OrderInMemory struct {
	mu sync.RWMutex
	db map[string]aggregate.Order
}

var _ Order = &OrderInMemory{}

func NewOrderInMemory() *OrderInMemory {
	return &OrderInMemory{
		mu: sync.RWMutex{},
		db: map[string]aggregate.Order{},
	}
}

func (o *OrderInMemory) Save(_ context.Context, order aggregate.Order) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.db[order.ID] = order
	return nil
}

func (o *OrderInMemory) Get(_ context.Context, s string) (aggregate.Order, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if order, ok := o.db[s]; ok {
		return order, nil
	}

	return aggregate.Order{}, errors.New("order not found")
}
