package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/go-asynq-loadtest/internal/domain"
	"github.com/lppduy/go-asynq-loadtest/internal/dto"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
	"github.com/lppduy/go-asynq-loadtest/internal/service"
)

// OrderHandler handles order HTTP requests
type OrderHandler struct {
	service service.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Create order
	order, err := h.service.CreateOrder(c.Request.Context(), req)
	if err != nil {
		log.Printf("Failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create order",
			Message: err.Error(),
		})
		return
	}

	// TODO: Enqueue background tasks
	// - payment:process (critical queue)
	// - inventory:update (high queue)
	// - email:confirmation (default queue)
	// - invoice:generate (default queue)
	// - analytics:track (low queue)
	// - warehouse:notify (low queue)

	log.Printf("✅ Order created: %s | Total: $%.2f | Items: %d", 
		order.ID, order.TotalAmount, len(order.Items))
	log.Printf("📋 Background tasks will be processed asynchronously")

	// Return response immediately (fast response!)
	c.JSON(http.StatusCreated, toOrderResponse(order))
}

// GetOrder handles GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	order, err := h.service.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		if err == repository.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Order not found",
				Message: fmt.Sprintf("Order %s does not exist", orderID),
			})
			return
		}

		log.Printf("Failed to get order: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get order",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

// ListOrders handles GET /api/v1/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Optional: filter by customer_id
	customerID := c.Query("customer_id")

	orders, err := h.service.ListOrders(c.Request.Context(), customerID)
	if err != nil {
		log.Printf("Failed to list orders: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list orders",
			Message: err.Error(),
		})
		return
	}

	// Convert to response DTOs
	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = toOrderResponse(order)
	}

	c.JSON(http.StatusOK, dto.OrderListResponse{
		Total:  len(orderResponses),
		Orders: orderResponses,
	})
}

// CancelOrder handles POST /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	order, err := h.service.CancelOrder(c.Request.Context(), orderID, req.Reason)
	if err != nil {
		if err == repository.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Order not found",
				Message: fmt.Sprintf("Order %s does not exist", orderID),
			})
			return
		}

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Cannot cancel order",
			Message: err.Error(),
		})
		return
	}

	// TODO: Enqueue refund task
	// - payment:refund (critical queue)

	log.Printf("⚠️ Order cancelled: %s | Reason: %s", order.ID, req.Reason)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Order cancelled successfully",
		Data:    toOrderResponse(order),
	})
}

// GetOrderStatus handles GET /api/v1/orders/:id/status
func (h *OrderHandler) GetOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	order, err := h.service.GetOrderStatus(c.Request.Context(), orderID)
	if err != nil {
		if err == repository.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Order not found",
				Message: fmt.Sprintf("Order %s does not exist", orderID),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get order status",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.OrderStatusResponse{
		OrderID:       order.ID,
		Status:        string(order.Status),
		PaymentStatus: string(order.PaymentStatus),
		UpdatedAt:     order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Helper function to convert domain.Order to dto.OrderResponse
func toOrderResponse(order *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		}
	}

	return dto.OrderResponse{
		ID:              order.ID,
		CustomerID:      order.CustomerID,
		CustomerEmail:   order.CustomerEmail,
		Items:           items,
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		Status:          string(order.Status),
		PaymentStatus:   string(order.PaymentStatus),
		PaymentMethod:   order.PaymentMethod,
		InvoiceURL:      order.InvoiceURL,
		TrackingNumber:  order.TrackingNumber,
		Notes:           order.Notes,
		CreatedAt:       order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
