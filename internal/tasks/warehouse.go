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
	TypeWarehouseNotify = "warehouse:notify"
)

// WarehousePayload represents the payload for warehouse notification
type WarehousePayload struct {
	OrderID         string `json:"order_id"`
	CustomerName    string `json:"customer_name"`
	ShippingAddress string `json:"shipping_address"`
	ItemCount       int    `json:"item_count"`
	Priority        string `json:"priority"` // standard, express, overnight
}

// NewWarehouseNotifyTask creates a new warehouse notification task
func NewWarehouseNotifyTask(orderID, customerName, shippingAddress string, itemCount int, priority string) (*asynq.Task, error) {
	payload, err := json.Marshal(WarehousePayload{
		OrderID:         orderID,
		CustomerName:    customerName,
		ShippingAddress: shippingAddress,
		ItemCount:       itemCount,
		Priority:        priority,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal warehouse payload: %w", err)
	}

	return asynq.NewTask(
		TypeWarehouseNotify,
		payload,
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Second),
		asynq.Queue("low"),            // Low priority
		asynq.ProcessIn(5*time.Second), // Notify after 5 seconds
	), nil
}

// HandleWarehouseNotifyTask notifies warehouse about new order
func HandleWarehouseNotifyTask(ctx context.Context, t *asynq.Task) error {
	var payload WarehousePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal warehouse payload: %w", err)
	}

	log.Printf("📦 [Warehouse] Notifying warehouse about order: %s", payload.OrderID)
	log.Printf("📦 [Warehouse] Customer: %s | Items: %d | Priority: %s", 
		payload.CustomerName, payload.ItemCount, payload.Priority)

	// Simulate warehouse notification (500ms)
	time.Sleep(500 * time.Millisecond)

	// Send notification to warehouse system
	if err := notifyWarehouseSystem(payload); err != nil {
		return fmt.Errorf("failed to notify warehouse: %w", err)
	}

	log.Printf("✅ [Warehouse] Notification sent for order: %s", payload.OrderID)
	return nil
}

// notifyWarehouseSystem sends notification to warehouse management system
func notifyWarehouseSystem(payload WarehousePayload) error {
	// In production: Call warehouse API or send message to queue
	// Example:
	// - REST API call to warehouse system
	// - Publish to Kafka/RabbitMQ
	// - Update warehouse database
	
	return nil
}
