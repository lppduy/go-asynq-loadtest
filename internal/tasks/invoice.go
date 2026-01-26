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
	TypeInvoiceGenerate = "invoice:generate"
)

// InvoicePayload represents the payload for invoice generation
type InvoicePayload struct {
	OrderID       string  `json:"order_id"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	TotalAmount   float64 `json:"total_amount"`
}

// NewInvoiceGenerateTask creates a new invoice generation task
func NewInvoiceGenerateTask(orderID, customerName, customerEmail string, totalAmount float64) (*asynq.Task, error) {
	payload, err := json.Marshal(InvoicePayload{
		OrderID:       orderID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		TotalAmount:   totalAmount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice payload: %w", err)
	}

	return asynq.NewTask(
		TypeInvoiceGenerate,
		payload,
		asynq.MaxRetry(3),
		asynq.Timeout(60*time.Second),  // PDF generation can take time
		asynq.Queue("default"),
		asynq.ProcessIn(5*time.Second), // Generate after 5 seconds
	), nil
}

// HandleInvoiceGenerateTask generates PDF invoice for order
func HandleInvoiceGenerateTask(ctx context.Context, t *asynq.Task) error {
	var payload InvoicePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal invoice payload: %w", err)
	}

	log.Printf("🧾 [Invoice] Generating invoice for order: %s", payload.OrderID)
	log.Printf("🧾 [Invoice] Customer: %s | Amount: $%.2f", payload.CustomerName, payload.TotalAmount)

	// Simulate PDF generation (3 seconds)
	time.Sleep(3 * time.Second)

	// Generate invoice PDF
	invoiceURL, err := generateInvoicePDF(payload)
	if err != nil {
		return fmt.Errorf("failed to generate invoice PDF: %w", err)
	}

	log.Printf("✅ [Invoice] Invoice generated: %s", invoiceURL)
	// In production: Update order with invoice URL
	// orderRepo.UpdateInvoiceURL(ctx, payload.OrderID, invoiceURL)

	return nil
}

// generateInvoicePDF generates PDF invoice and uploads to storage
func generateInvoicePDF(payload InvoicePayload) (string, error) {
	// In production:
	// 1. Generate PDF using library (e.g., go-pdf, wkhtmltopdf)
	// 2. Upload to S3/GCS
	// 3. Return public URL
	
	// Simulate PDF generation
	invoiceURL := fmt.Sprintf("https://storage.example.com/invoices/%s.pdf", payload.OrderID)
	return invoiceURL, nil
}
