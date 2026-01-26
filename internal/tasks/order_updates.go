package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/lppduy/go-asynq-loadtest/internal/domain"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
)

func updateOrder(ctx context.Context, repo repository.OrderRepository, orderID string, mutate func(o *domain.Order)) error {
	order, err := repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	mutate(order)
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = time.Now()
	}

	if err := repo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order %s: %w", orderID, err)
	}
	return nil
}

