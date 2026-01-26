package domain

import "time"

// OrderStatus represents the current state of an order
type OrderStatus string

const (
	OrderStatusPending           OrderStatus = "pending"
	OrderStatusPaymentProcessing OrderStatus = "payment_processing"
	OrderStatusPaymentFailed     OrderStatus = "payment_failed"
	OrderStatusConfirmed         OrderStatus = "confirmed"
	OrderStatusProcessing        OrderStatus = "processing"
	OrderStatusShipped           OrderStatus = "shipped"
	OrderStatusDelivered         OrderStatus = "delivered"
	OrderStatusCancelled         OrderStatus = "cancelled"
)

// PaymentStatus represents the payment state
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Order represents an order in the system
type Order struct {
	ID              string        `json:"id"`
	CustomerID      string        `json:"customer_id"`
	CustomerEmail   string        `json:"customer_email"`
	Items           []OrderItem   `json:"items"`
	TotalAmount     float64       `json:"total_amount"`
	ShippingAddress Address       `json:"shipping_address"`
	Status          OrderStatus   `json:"status"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	PaymentMethod   string        `json:"payment_method"`
	InvoiceURL      string        `json:"invoice_url,omitempty"`
	TrackingNumber  string        `json:"tracking_number,omitempty"`
	Notes           string        `json:"notes,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// OrderItem represents a product in an order
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

// Address represents a shipping/billing address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// CalculateTotal calculates the total amount for all items
func (o *Order) CalculateTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.Subtotal
	}
	return total
}

// CanCancel checks if order can be cancelled
func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPending ||
		o.Status == OrderStatusPaymentProcessing ||
		o.Status == OrderStatusConfirmed
}

// Cancel cancels the order
func (o *Order) Cancel() {
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
}

// UpdatePaymentStatus updates payment status and order status accordingly
func (o *Order) UpdatePaymentStatus(status PaymentStatus) {
	o.PaymentStatus = status
	o.UpdatedAt = time.Now()

	switch status {
	case PaymentStatusCompleted:
		o.Status = OrderStatusConfirmed
	case PaymentStatusFailed:
		o.Status = OrderStatusPaymentFailed
	}
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}
