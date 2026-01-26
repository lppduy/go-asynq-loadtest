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
	TypeAnalyticsTrack = "analytics:track"
)

// AnalyticsPayload represents the payload for analytics tracking
type AnalyticsPayload struct {
	OrderID      string  `json:"order_id"`
	CustomerID   string  `json:"customer_id"`
	TotalAmount  float64 `json:"total_amount"`
	ItemCount    int     `json:"item_count"`
	PaymentMethod string `json:"payment_method"`
	CreatedAt    string  `json:"created_at"`
}

// NewAnalyticsTrackTask creates a new analytics tracking task
func NewAnalyticsTrackTask(orderID, customerID string, totalAmount float64, itemCount int, paymentMethod string) (*asynq.Task, error) {
	payload, err := json.Marshal(AnalyticsPayload{
		OrderID:       orderID,
		CustomerID:    customerID,
		TotalAmount:   totalAmount,
		ItemCount:     itemCount,
		PaymentMethod: paymentMethod,
		CreatedAt:     time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analytics payload: %w", err)
	}

	return asynq.NewTask(
		TypeAnalyticsTrack,
		payload,
		asynq.MaxRetry(2),             // Analytics can fail without blocking order
		asynq.Timeout(10*time.Second),
		asynq.Queue("low"),            // Low priority
		asynq.ProcessIn(10*time.Second), // Track after 10 seconds
	), nil
}

// HandleAnalyticsTrackTask tracks order analytics
func HandleAnalyticsTrackTask(ctx context.Context, t *asynq.Task) error {
	var payload AnalyticsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal analytics payload: %w", err)
	}

	log.Printf("📊 [Analytics] Tracking order: %s", payload.OrderID)
	log.Printf("📊 [Analytics] Customer: %s | Amount: $%.2f | Items: %d", 
		payload.CustomerID, payload.TotalAmount, payload.ItemCount)

	// Simulate analytics tracking (200ms)
	time.Sleep(200 * time.Millisecond)

	// Send to analytics service
	if err := sendToAnalytics(payload); err != nil {
		// Log but don't fail - analytics is not critical
		log.Printf("⚠️  [Analytics] Failed to send data: %v", err)
		return nil // Don't retry
	}

	log.Printf("✅ [Analytics] Event tracked for order: %s", payload.OrderID)
	return nil
}

// sendToAnalytics sends data to analytics service
func sendToAnalytics(payload AnalyticsPayload) error {
	// In production: Send to Google Analytics, Mixpanel, Segment, etc.
	// Example:
	// analytics.Track("order_created", map[string]interface{}{
	//     "order_id": payload.OrderID,
	//     "revenue": payload.TotalAmount,
	//     "item_count": payload.ItemCount,
	// })
	
	return nil
}
