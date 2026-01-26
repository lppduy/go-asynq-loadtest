package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/lppduy/go-asynq-loadtest/internal/domain"
)

// GormOrderRepository implements OrderRepository using GORM
type GormOrderRepository struct {
	db *gorm.DB
}

// NewGormOrderRepository creates a new GORM-based repository
func NewGormOrderRepository(db *gorm.DB) OrderRepository {
	return &GormOrderRepository{db: db}
}

// Create adds a new order
func (r *GormOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	model, err := domain.FromOrder(order)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves an order by ID
func (r *GormOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	var model domain.OrderModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return model.ToOrder()
}

// FindByCustomerID retrieves all orders for a customer
func (r *GormOrderRepository) FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error) {
	var models []domain.OrderModel
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(models))
	for _, model := range models {
		order, err := model.ToOrder()
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// Update updates an existing order
func (r *GormOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	model, err := domain.FromOrder(order)
	if err != nil {
		return err
	}

	result := r.db.WithContext(ctx).
		Model(&domain.OrderModel{}).
		Where("id = ?", order.ID).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}

// Delete removes an order by ID (soft delete)
func (r *GormOrderRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.OrderModel{}, "id = ?", id)
	
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}

// FindAll retrieves all orders
func (r *GormOrderRepository) FindAll(ctx context.Context) ([]*domain.Order, error) {
	var models []domain.OrderModel
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(models))
	for _, model := range models {
		order, err := model.ToOrder()
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}
