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
	TypeInventoryUpdate = "inventory:update"
)

// InventoryItem represents an item to update in inventory
type InventoryItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// InventoryPayload represents the payload for inventory update
type InventoryPayload struct {
	OrderID string          `json:"order_id"`
	Items   []InventoryItem `json:"items"`
}

// NewInventoryUpdateTask creates a new inventory update task
func NewInventoryUpdateTask(orderID string, items []InventoryItem) (*asynq.Task, error) {
	payload, err := json.Marshal(InventoryPayload{
		OrderID: orderID,
		Items:   items,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inventory payload: %w", err)
	}

	return asynq.NewTask(
		TypeInventoryUpdate,
		payload,
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Second),
		asynq.Queue("high"),            // High priority
		asynq.ProcessIn(1*time.Second), // Process quickly
	), nil
}

// HandleInventoryUpdateTask updates inventory for order items
func HandleInventoryUpdateTask(ctx context.Context, t *asynq.Task) error {
	var payload InventoryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal inventory payload: %w", err)
	}

	log.Printf("📦 [Inventory] Updating inventory for order: %s", payload.OrderID)
	log.Printf("📦 [Inventory] Items to update: %d", len(payload.Items))

	// Simulate inventory update (500ms)
	time.Sleep(500 * time.Millisecond)

	// Process each item
	for _, item := range payload.Items {
		if err := updateInventoryItem(item); err != nil {
			return fmt.Errorf("failed to update inventory for product %s: %w", item.ProductID, err)
		}
		log.Printf("📦 [Inventory] Updated: %s (qty: %d)", item.ProductID, item.Quantity)
	}

	log.Printf("✅ [Inventory] All items updated for order: %s", payload.OrderID)
	return nil
}

// updateInventoryItem updates a single inventory item
func updateInventoryItem(item InventoryItem) error {
	// In production: Update database inventory table
	// Example: inventoryRepo.DecrementStock(item.ProductID, item.Quantity)
	
	// Simulate database update
	time.Sleep(100 * time.Millisecond)
	return nil
}
