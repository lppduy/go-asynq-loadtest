package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeEmailConfirmation = "email:confirmation"
)

// EmailPayload represents the payload for email sending
type EmailPayload struct {
	OrderID       string `json:"order_id"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
	TotalAmount   float64 `json:"total_amount"`
}

// NewEmailConfirmationTask creates a new email confirmation task
func NewEmailConfirmationTask(orderID, customerEmail, customerName string, totalAmount float64) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailPayload{
		OrderID:       orderID,
		CustomerEmail: customerEmail,
		CustomerName:  customerName,
		TotalAmount:   totalAmount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email payload: %w", err)
	}

	return asynq.NewTask(
		TypeEmailConfirmation,
		payload,
		asynq.MaxRetry(5),              // Email can retry more
		asynq.Timeout(20*time.Second),
		asynq.Queue("default"),         // Default queue
		asynq.ProcessIn(3*time.Second), // Send after 3 seconds
	), nil
}

// HandleEmailConfirmationTask sends order confirmation email
func HandleEmailConfirmationTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %w", err)
	}

	log.Printf("📧 [Email] Sending confirmation to: %s", payload.CustomerEmail)
	log.Printf("📧 [Email] Order: %s | Amount: $%.2f", payload.OrderID, payload.TotalAmount)

	// Simulate email sending (1 second)
	time.Sleep(1 * time.Second)

	// Simulate email service API call
	// In production: Use SendGrid, AWS SES, Mailgun, etc.
	if err := sendEmail(payload); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ [Email] Confirmation sent successfully to: %s", payload.CustomerEmail)
	return nil
}

// sendEmail simulates sending email via email service
func sendEmail(payload EmailPayload) error {
	// In production: Call actual email service API
	// Example: SendGrid, AWS SES, Mailgun
	emailBody := fmt.Sprintf(`
		Dear Customer,
		
		Your order %s has been confirmed!
		Total Amount: $%.2f
		
		Thank you for your purchase!
	`, payload.OrderID, payload.TotalAmount)

	_ = emailBody // Use the email body
	return nil
}
