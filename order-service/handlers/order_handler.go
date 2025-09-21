package handlers

import (
	"net/http"
	"order-service/dto/request"
	"order-service/dto/response"
	"order-service/entities"
	"order-service/helpers"
	"order-service/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
	req := new(request.CreateOrderRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   err.Error(),
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   helpers.TranslateValidationErr(err).Error(),
		})
	}

	order := entities.Order{
		ProductID: req.ProductID,
		Qty:       req.Qty,
		Status:    "pending",
	}

	_, err := h.orderService.Create(order)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusInternalServerError),
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusAccepted, response.BaseResponse{
		Status:  true,
		Message: http.StatusText(http.StatusAccepted),
		Data: map[string]interface{}{
			"product_id": order.ProductID,
			"qty":        order.Qty,
			"status":     order.Status,
		},
	})
}

func (h *OrderHandler) FindAllOrders(c echo.Context) error {
	orders, err := h.orderService.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusInternalServerError),
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.BaseResponse{
		Status:  false,
		Message: http.StatusText(http.StatusOK),
		Data:    orders,
	})
}

func (h *OrderHandler) FindOrderByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   err.Error(),
		})
	}

	order, err := h.orderService.FindByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusNotFound),
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.BaseResponse{
		Status:  false,
		Message: http.StatusText(http.StatusOK),
		Data:    order,
	})
}

func (h *OrderHandler) FindOrdersByProductID(c echo.Context) error {
	productID, err := strconv.ParseUint(c.Param("productID"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   err.Error(),
		})
	}

	orders, err := h.orderService.FindByProductID(uint(productID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusInternalServerError),
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.BaseResponse{
		Status:  false,
		Message: http.StatusText(http.StatusOK),
		Data:    orders,
	})
}

func (h *OrderHandler) UpdateOrder(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   err.Error(),
		})
	}

	req := new(request.UpdateOrderRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   err.Error(),
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   helpers.TranslateValidationErr(err).Error(),
		})
	}

	order := entities.Order{
		ID:     uint(id),
		Status: req.Status,
	}

	updatedOrder, err := h.orderService.Update(order)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusInternalServerError),
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.BaseResponse{
		Status:  false,
		Message: http.StatusText(http.StatusOK),
		Data:    updatedOrder,
	})
}

func (h *OrderHandler) DeleteOrder(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusBadRequest),
			Error:   err.Error(),
		})
	}

	if err := h.orderService.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, response.BaseResponse{
			Status:  false,
			Message: http.StatusText(http.StatusInternalServerError),
			Error:   err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}
