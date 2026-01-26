package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lppduy/go-asynq-loadtest/internal/domain"
	"github.com/lppduy/go-asynq-loadtest/internal/dto"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
)

// OrderService defines business logic for orders
type OrderService interface {
	CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	ListOrders(ctx context.Context, customerID string) ([]*domain.Order, error)
	CancelOrder(ctx context.Context, id string, reason string) (*domain.Order, error)
	GetOrderStatus(ctx context.Context, id string) (*domain.Order, error)
}

type orderService struct {
	repo repository.OrderRepository
}

// NewOrderService creates a new order service
func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

// CreateOrder creates a new order
func (s *orderService) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*domain.Order, error) {
	// Generate order ID
	orderID := generateOrderID()

	// Convert items and calculate totals
	items := make([]domain.OrderItem, len(req.Items))
	totalAmount := 0.0

	for i, item := range req.Items {
		subtotal := item.UnitPrice * float64(item.Quantity)
		items[i] = domain.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    subtotal,
		}
		totalAmount += subtotal
	}

	// Create order
	now := time.Now()
	order := &domain.Order{
		ID:              orderID,
		CustomerID:      req.CustomerID,
		CustomerEmail:   req.CustomerEmail,
		Items:           items,
		TotalAmount:     totalAmount,
		ShippingAddress: req.ShippingAddress,
		Status:          domain.OrderStatusPending,
		PaymentStatus:   domain.PaymentStatusPending,
		PaymentMethod:   req.PaymentMethod,
		Notes:           req.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Save to repository
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Note: Background jobs will be enqueued by the handler
	// - payment:process
	// - inventory:update
	// - email:confirmation
	// - invoice:generate
	// - analytics:track
	// - warehouse:notify

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *orderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

// ListOrders retrieves all orders for a customer
func (s *orderService) ListOrders(ctx context.Context, customerID string) ([]*domain.Order, error) {
	if customerID == "" {
		// Return all orders (for demo purposes)
		return s.repo.FindAll(ctx)
	}
	return s.repo.FindByCustomerID(ctx, customerID)
}

// CancelOrder cancels an order
func (s *orderService) CancelOrder(ctx context.Context, id string, reason string) (*domain.Order, error) {
	// Get order
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if can cancel
	if !order.CanCancel() {
		return nil, fmt.Errorf("order cannot be cancelled (current status: %s)", order.Status)
	}

	// Cancel order
	order.Cancel()
	order.Notes = fmt.Sprintf("Cancelled: %s", reason)

	// Update repository
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Note: Background job to process refund will be enqueued by handler
	// - payment:refund

	return order, nil
}

// GetOrderStatus retrieves order status
func (s *orderService) GetOrderStatus(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

// Helper function to generate order ID
func generateOrderID() string {
	return fmt.Sprintf("ORD-%s", uuid.New().String()[:8])
}
