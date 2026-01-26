package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/lppduy/go-asynq-loadtest/internal/domain"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
)

// Task type constants
const (
	TypePaymentProcess = "payment:process"
)

// PaymentPayload represents the payload for payment processing
type PaymentPayload struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
}

// NewPaymentProcessTask creates a new payment processing task
func NewPaymentProcessTask(orderID string, amount float64, paymentMethod string) (*asynq.Task, error) {
	payload, err := json.Marshal(PaymentPayload{
		OrderID:       orderID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment payload: %w", err)
	}

	// Create task with options
	return asynq.NewTask(
		TypePaymentProcess,
		payload,
		asynq.MaxRetry(3),                    // Retry up to 3 times
		asynq.Timeout(30*time.Second),        // Task timeout
		asynq.Queue("critical"),              // Use critical queue
		asynq.ProcessIn(2*time.Second),       // Process after 2 seconds (simulate delay)
	), nil
}

// NewPaymentProcessHandler returns a handler that also updates order status in PostgreSQL.
func NewPaymentProcessHandler(orderRepo repository.OrderRepository) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload PaymentPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payment payload: %w", err)
		}

		// Mark payment as processing immediately so orders don't remain "pending"
		_ = updateOrder(ctx, orderRepo, payload.OrderID, func(o *domain.Order) {
			o.UpdateStatus(domain.OrderStatusPaymentProcessing)
			o.UpdatePaymentStatus(domain.PaymentStatusProcessing)
		})

		log.Printf("💳 [Payment] Processing payment for order: %s", payload.OrderID)
		log.Printf("💳 [Payment] Amount: $%.2f | Method: %s", payload.Amount, payload.PaymentMethod)

		// Simulate payment processing (2 seconds)
		time.Sleep(2 * time.Second)

		// Simulate payment gateway API call
		success := simulatePaymentGateway(payload)
		if !success {
			_ = updateOrder(ctx, orderRepo, payload.OrderID, func(o *domain.Order) {
				o.UpdatePaymentStatus(domain.PaymentStatusFailed)
			})
			return fmt.Errorf("payment failed for order %s", payload.OrderID)
		}

		// Persist success into PostgreSQL
		if err := updateOrder(ctx, orderRepo, payload.OrderID, func(o *domain.Order) {
			o.UpdatePaymentStatus(domain.PaymentStatusCompleted)
			o.UpdateStatus(domain.OrderStatusConfirmed)
		}); err != nil {
			return err
		}

		log.Printf("✅ [Payment] Payment processed successfully for order: %s", payload.OrderID)
		return nil
	}
}

// HandlePaymentProcessTask processes payment for an order
func HandlePaymentProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload PaymentPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payment payload: %w", err)
	}

	log.Printf("💳 [Payment] Processing payment for order: %s", payload.OrderID)
	log.Printf("💳 [Payment] Amount: $%.2f | Method: %s", payload.Amount, payload.PaymentMethod)

	// Simulate payment processing (2 seconds)
	time.Sleep(2 * time.Second)

	// Simulate payment gateway API call
	// In production: Call Stripe, PayPal, etc.
	success := simulatePaymentGateway(payload)

	if !success {
		return fmt.Errorf("payment failed for order %s", payload.OrderID)
	}

	log.Printf("✅ [Payment] Payment processed successfully for order: %s", payload.OrderID)
	// In production: Update order payment status in database
	// orderRepo.UpdatePaymentStatus(ctx, payload.OrderID, "paid")

	return nil
}

// simulatePaymentGateway simulates external payment gateway
func simulatePaymentGateway(payload PaymentPayload) bool {
	// Simulate 95% success rate
	// In production: Make actual API call to payment gateway
	return true
}
