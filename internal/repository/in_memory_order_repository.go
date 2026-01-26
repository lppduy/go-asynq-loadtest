package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/lppduy/go-asynq-loadtest/internal/domain"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

// OrderRepository defines methods for order data access
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context) ([]*domain.Order, error)
}

// InMemoryOrderRepository implements OrderRepository using in-memory storage
type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

// NewInMemoryOrderRepository creates a new in-memory repository
func NewInMemoryOrderRepository() OrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*domain.Order),
	}
}

// Create adds a new order
func (r *InMemoryOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order
	return nil
}

// FindByID retrieves an order by ID
func (r *InMemoryOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}

	return order, nil
}

// FindByCustomerID retrieves all orders for a customer
func (r *InMemoryOrderRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []*domain.Order
	for _, order := range r.orders {
		if order.CustomerID == customerID {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// Update updates an existing order
func (r *InMemoryOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return ErrOrderNotFound
	}

	r.orders[order.ID] = order
	return nil
}

// Delete removes an order by ID
func (r *InMemoryOrderRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[id]; !exists {
		return ErrOrderNotFound
	}

	delete(r.orders, id)
	return nil
}

// FindAll retrieves all orders
func (r *InMemoryOrderRepository) FindAll(ctx context.Context) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []*domain.Order
	for _, order := range r.orders {
		orders = append(orders, order)
	}

	return orders, nil
}
