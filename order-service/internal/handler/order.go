// internal/handler/order.go
package handler

import (
	"net/http"

	"order-service/internal/service"

	"github.com/labstack/echo/v4"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
	userID, _ := c.Get("user_id").(int64) // можно передавать через контекст в middleware

	type Request struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	}

	req := new(Request)
	if err := c.Bind(req); err != nil {
		return echo.ErrBadRequest
	}

	if req.Quantity <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "quantity must be positive")
	}

	if err := h.orderService.CreateOrder(c.Request().Context(), userID, req.ProductID, req.Quantity); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Order created"})
}
