package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// OrderModel represents the order table in database (GORM model)
type OrderModel struct {
	ID              string         `gorm:"primaryKey;type:varchar(50)"`
	CustomerID      string         `gorm:"type:varchar(100);not null;index"`
	CustomerEmail   string         `gorm:"type:varchar(255);not null"`
	ItemsJSON       string         `gorm:"type:text;not null"` // JSON string of items
	TotalAmount     float64        `gorm:"type:decimal(10,2);not null"`
	AddressJSON     string         `gorm:"type:text;not null"` // JSON string of address
	Status          string         `gorm:"type:varchar(50);not null;default:'pending';index"`
	PaymentStatus   string         `gorm:"type:varchar(50);not null;default:'pending'"`
	PaymentMethod   string         `gorm:"type:varchar(50);not null"`
	InvoiceURL      string         `gorm:"type:varchar(500)"`
	TrackingNumber  string         `gorm:"type:varchar(100)"`
	Notes           string         `gorm:"type:text"`
	CreatedAt       time.Time      `gorm:"not null;index"`
	UpdatedAt       time.Time      `gorm:"not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"` // Soft delete support
}

// TableName overrides the table name
func (OrderModel) TableName() string {
	return "orders"
}

// ToOrder converts OrderModel to domain.Order
func (m *OrderModel) ToOrder() (*Order, error) {
	var items []OrderItem
	if err := json.Unmarshal([]byte(m.ItemsJSON), &items); err != nil {
		return nil, err
	}

	var address Address
	if err := json.Unmarshal([]byte(m.AddressJSON), &address); err != nil {
		return nil, err
	}

	return &Order{
		ID:              m.ID,
		CustomerID:      m.CustomerID,
		CustomerEmail:   m.CustomerEmail,
		Items:           items,
		TotalAmount:     m.TotalAmount,
		ShippingAddress: address,
		Status:          OrderStatus(m.Status),
		PaymentStatus:   PaymentStatus(m.PaymentStatus),
		PaymentMethod:   m.PaymentMethod,
		InvoiceURL:      m.InvoiceURL,
		TrackingNumber:  m.TrackingNumber,
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}, nil
}

// FromOrder converts domain.Order to OrderModel
func FromOrder(order *Order) (*OrderModel, error) {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return nil, err
	}

	addressJSON, err := json.Marshal(order.ShippingAddress)
	if err != nil {
		return nil, err
	}

	return &OrderModel{
		ID:             order.ID,
		CustomerID:     order.CustomerID,
		CustomerEmail:  order.CustomerEmail,
		ItemsJSON:      string(itemsJSON),
		TotalAmount:    order.TotalAmount,
		AddressJSON:    string(addressJSON),
		Status:         string(order.Status),
		PaymentStatus:  string(order.PaymentStatus),
		PaymentMethod:  order.PaymentMethod,
		InvoiceURL:     order.InvoiceURL,
		TrackingNumber: order.TrackingNumber,
		Notes:          order.Notes,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}, nil
}
