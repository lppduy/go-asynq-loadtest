package dto

import "github.com/lppduy/go-asynq-loadtest/internal/domain"

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID      string                `json:"customer_id" binding:"required"`
	CustomerEmail   string                `json:"customer_email" binding:"required,email"`
	Items           []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
	ShippingAddress domain.Address        `json:"shipping_address" binding:"required"`
	PaymentMethod   string                `json:"payment_method" binding:"required,oneof=credit_card debit_card bank_transfer"`
	Notes           string                `json:"notes"`
}

// CreateOrderItemRequest represents an item in the order creation request
type CreateOrderItemRequest struct {
	ProductID   string  `json:"product_id" binding:"required"`
	ProductName string  `json:"product_name" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,gt=0"`
}

// OrderResponse represents the response for an order
type OrderResponse struct {
	ID              string                `json:"id"`
	CustomerID      string                `json:"customer_id"`
	CustomerEmail   string                `json:"customer_email"`
	Items           []OrderItemResponse   `json:"items"`
	TotalAmount     float64               `json:"total_amount"`
	ShippingAddress domain.Address        `json:"shipping_address"`
	Status          string                `json:"status"`
	PaymentStatus   string                `json:"payment_status"`
	PaymentMethod   string                `json:"payment_method"`
	InvoiceURL      string                `json:"invoice_url,omitempty"`
	TrackingNumber  string                `json:"tracking_number,omitempty"`
	Notes           string                `json:"notes,omitempty"`
	CreatedAt       string                `json:"created_at"`
	UpdatedAt       string                `json:"updated_at"`
}

// OrderItemResponse represents an item in the order response
type OrderItemResponse struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

// OrderListResponse represents the response for list of orders
type OrderListResponse struct {
	Total  int             `json:"total"`
	Orders []OrderResponse `json:"orders"`
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required,min=5"`
}

// OrderStatusResponse represents the response for order status
type OrderStatusResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	UpdatedAt     string `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
