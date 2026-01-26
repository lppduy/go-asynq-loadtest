package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/lppduy/go-asynq-loadtest/internal/domain"
	"github.com/lppduy/go-asynq-loadtest/internal/dto"
	"github.com/lppduy/go-asynq-loadtest/internal/repository"
	"github.com/lppduy/go-asynq-loadtest/internal/service"
	"github.com/lppduy/go-asynq-loadtest/internal/tasks"
)

// OrderHandler handles order HTTP requests
type OrderHandler struct {
	service     service.OrderService
	asynqClient *asynq.Client
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service service.OrderService, asynqClient *asynq.Client) *OrderHandler {
	return &OrderHandler{
		service:     service,
		asynqClient: asynqClient,
	}
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

	// Enqueue background tasks (non-blocking, fast response)
	go h.enqueueOrderTasks(order)

	log.Printf("✅ Order created: %s | Total: $%.2f | Items: %d", 
		order.ID, order.TotalAmount, len(order.Items))
	log.Printf("📋 Background tasks enqueued asynchronously")

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

// enqueueOrderTasks enqueues all background tasks for order processing
func (h *OrderHandler) enqueueOrderTasks(order *domain.Order) {
	// 1. Payment Processing (Critical Queue - highest priority)
	paymentTask, err := tasks.NewPaymentProcessTask(order.ID, order.TotalAmount, order.PaymentMethod)
	if err != nil {
		log.Printf("❌ Failed to create payment task: %v", err)
	} else {
		if _, err := h.asynqClient.Enqueue(paymentTask); err != nil {
			log.Printf("❌ Failed to enqueue payment task: %v", err)
		} else {
			log.Printf("📤 [Enqueued] Payment task for order: %s", order.ID)
		}
	}

	// 2. Inventory Update (High Queue)
	inventoryItems := make([]tasks.InventoryItem, len(order.Items))
	for i, item := range order.Items {
		inventoryItems[i] = tasks.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	inventoryTask, err := tasks.NewInventoryUpdateTask(order.ID, inventoryItems)
	if err != nil {
		log.Printf("❌ Failed to create inventory task: %v", err)
	} else {
		if _, err := h.asynqClient.Enqueue(inventoryTask); err != nil {
			log.Printf("❌ Failed to enqueue inventory task: %v", err)
		} else {
			log.Printf("📤 [Enqueued] Inventory task for order: %s", order.ID)
		}
	}

	// 3. Email Confirmation (Default Queue)
	emailTask, err := tasks.NewEmailConfirmationTask(
		order.ID,
		order.CustomerEmail,
		order.CustomerID, // Using customer ID as name for demo
		order.TotalAmount,
	)
	if err != nil {
		log.Printf("❌ Failed to create email task: %v", err)
	} else {
		if _, err := h.asynqClient.Enqueue(emailTask); err != nil {
			log.Printf("❌ Failed to enqueue email task: %v", err)
		} else {
			log.Printf("📤 [Enqueued] Email task for order: %s", order.ID)
		}
	}

	// 4. Invoice Generation (Default Queue)
	invoiceTask, err := tasks.NewInvoiceGenerateTask(
		order.ID,
		order.CustomerID,
		order.CustomerEmail,
		order.TotalAmount,
	)
	if err != nil {
		log.Printf("❌ Failed to create invoice task: %v", err)
	} else {
		if _, err := h.asynqClient.Enqueue(invoiceTask); err != nil {
			log.Printf("❌ Failed to enqueue invoice task: %v", err)
		} else {
			log.Printf("📤 [Enqueued] Invoice task for order: %s", order.ID)
		}
	}

	// 5. Analytics Tracking (Low Queue)
	analyticsTask, err := tasks.NewAnalyticsTrackTask(
		order.ID,
		order.CustomerID,
		order.TotalAmount,
		len(order.Items),
		order.PaymentMethod,
	)
	if err != nil {
		log.Printf("❌ Failed to create analytics task: %v", err)
	} else {
		if _, err := h.asynqClient.Enqueue(analyticsTask); err != nil {
			log.Printf("❌ Failed to enqueue analytics task: %v", err)
		} else {
			log.Printf("📤 [Enqueued] Analytics task for order: %s", order.ID)
		}
	}

	// 6. Warehouse Notification (Low Queue)
	shippingAddr := fmt.Sprintf("%s, %s, %s %s", 
		order.ShippingAddress.Street,
		order.ShippingAddress.City,
		order.ShippingAddress.State,
		order.ShippingAddress.PostalCode,
	)
	warehouseTask, err := tasks.NewWarehouseNotifyTask(
		order.ID,
		order.CustomerID,
		shippingAddr,
		len(order.Items),
		"standard", // Default priority
	)
	if err != nil {
		log.Printf("❌ Failed to create warehouse task: %v", err)
	} else {
		if _, err := h.asynqClient.Enqueue(warehouseTask); err != nil {
			log.Printf("❌ Failed to enqueue warehouse task: %v", err)
		} else {
			log.Printf("📤 [Enqueued] Warehouse task for order: %s", order.ID)
		}
	}

	log.Printf("✅ All background tasks enqueued for order: %s", order.ID)
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
