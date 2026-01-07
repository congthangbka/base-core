package handler

import (
	"llm-aggregator/internal/common"
	"llm-aggregator/internal/modules/order/dto"
	"llm-aggregator/internal/modules/order/service"
	"llm-aggregator/internal/modules/order/validator"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service   service.OrderService
	validator *validator.OrderValidator
}

func NewOrderHandler(service service.OrderService, validator *validator.OrderValidator) *OrderHandler {
	return &OrderHandler{
		service:   service,
		validator: validator,
	}
}

// Create handles POST /orders
// @Summary     Create a new order
// @Description Create a new order for a user
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       order body     dto.CreateOrderRequest true "Order data"
// @Success     201   {object} common.Response{data=dto.OrderResponse}
// @Failure     400   {object} common.Response
// @Failure     404   {object} common.Response
// @Failure     500   {object} common.Response
// @Router      /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.validator.ValidateCreateRequest(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	order, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondCreated(c, order)
}

// GetAll handles GET /orders
// @Summary     Get all orders
// @Description Get a paginated list of orders with optional filters
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       page        query    int    false "Page number" default(1)
// @Param       limit       query    int    false "Items per page" default(10)
// @Param       userId      query    string false "Filter by user ID"
// @Param       productName query    string false "Filter by product name"
// @Param       status      query    int    false "Filter by status (1=pending, 2=completed, 3=cancelled)"
// @Success     200         {object} common.Response{data=dto.OrderPagingResponse}
// @Failure     500         {object} common.Response
// @Router      /orders [get]
func (h *OrderHandler) GetAll(c *gin.Context) {
	var req dto.OrderPagingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	orders, err := h.service.GetAll(c.Request.Context(), &req)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccessWithPagination(c, orders.Data, orders.Page, orders.Limit, orders.Total)
}

// GetByID handles GET /orders/:id
// @Summary     Get order by ID
// @Description Get a specific order by its ID
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Order ID"
// @Success     200  {object} common.Response{data=dto.OrderResponse}
// @Failure     404  {object} common.Response
// @Failure     500  {object} common.Response
// @Router      /orders/{id} [get]
func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.RespondBadRequest(c, "Order ID is required")
		return
	}

	order, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccess(c, order)
}

// Update handles PUT /orders/:id
// @Summary     Update an order
// @Description Update an existing order
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       id     path     string                true "Order ID"
// @Param       order  body     dto.UpdateOrderRequest true "Order data"
// @Success     200    {object} common.Response
// @Failure     400    {object} common.Response
// @Failure     404    {object} common.Response
// @Failure     500    {object} common.Response
// @Router      /orders/{id} [put]
func (h *OrderHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.RespondBadRequest(c, "Order ID is required")
		return
	}

	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.validator.ValidateUpdateRequest(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.service.Update(c.Request.Context(), id, &req); err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccess(c, nil)
}

// Delete handles DELETE /orders/:id
// @Summary     Delete an order
// @Description Delete an order by its ID
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       id   path     string true "Order ID"
// @Success     200  {object} common.Response
// @Failure     404  {object} common.Response
// @Failure     500  {object} common.Response
// @Router      /orders/{id} [delete]
func (h *OrderHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.RespondBadRequest(c, "Order ID is required")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccess(c, nil)
}

// GetByUserID handles GET /orders/user/:userId
// @Summary     Get orders by user ID
// @Description Get all orders for a specific user
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       userId path     string true "User ID"
// @Param       page   query    int    false "Page number" default(1)
// @Param       limit  query    int    false "Items per page" default(10)
// @Success     200    {object} common.Response{data=dto.OrderPagingResponse}
// @Failure     404    {object} common.Response
// @Failure     500    {object} common.Response
// @Router      /orders/user/{userId} [get]
func (h *OrderHandler) GetByUserID(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		common.RespondBadRequest(c, "User ID is required")
		return
	}

	// Get pagination params
	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		var req dto.OrderPagingRequest
		if err := c.ShouldBindQuery(&req); err == nil {
			if req.Page > 0 {
				page = req.Page
			}
			if req.Limit > 0 {
				limit = req.Limit
			}
		}
	}

	orders, err := h.service.GetByUserID(c.Request.Context(), userID, page, limit)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccessWithPagination(c, orders.Data, orders.Page, orders.Limit, orders.Total)
}

