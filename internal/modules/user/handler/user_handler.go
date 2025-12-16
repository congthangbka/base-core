package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/example/clean-architecture/internal/common"
	"github.com/example/clean-architecture/internal/modules/user/dto"
	"github.com/example/clean-architecture/internal/modules/user/service"
	"github.com/example/clean-architecture/internal/modules/user/validator"
)

type UserHandler struct {
	service   service.UserService
	validator *validator.UserValidator
}

func NewUserHandler(service service.UserService, validator *validator.UserValidator) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: validator,
	}
}

// @Summary     Create a new user
// @Description Create a new user with name and email
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user body     dto.CreateUserRequest true "User data"
// @Success     201  {object} common.SuccessResponseDoc{data=dto.UserResponse}
// @Failure     400  {object} common.ErrorResponseDoc "Bad Request - Possible error codes: BAD_REQUEST, VALIDATION_ERROR, EMAIL_EXISTS, USER_ALREADY_EXISTS"
// @Failure     500  {object} common.ErrorResponseDoc "Internal Server Error - Error code: INTERNAL_ERROR"
// @Router      /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.validator.ValidateCreate(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	user, err := h.service.Create(ctx, &req)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondCreated(c, user)
}

// @Summary     Get user by ID
// @Description Get a user by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id   path     string true "User ID"
// @Success     200  {object} common.SuccessResponseDoc{data=dto.UserResponse}
// @Failure     404  {object} common.ErrorResponseDoc "Not Found - Error code: USER_NOT_FOUND"
// @Failure     500  {object} common.ErrorResponseDoc "Internal Server Error - Error code: INTERNAL_ERROR"
// @Router      /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		common.RespondBadRequest(c, "ID is required")
		return
	}

	user, err := h.service.GetByID(ctx, id)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccess(c, user)
}

// @Summary     Get all users
// @Description Get all users with pagination and filters
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       page  query    int    false "Page number" default(1)
// @Param       limit query    int    false "Items per page" default(20)
// @Param       name  query    string false "Filter by name"
// @Param       email query    string false "Filter by email"
// @Success     200   {object} common.SuccessResponseWithPaginationDoc{data=[]dto.UserResponse}
// @Failure     400   {object} common.ErrorResponseDoc "Bad Request - Possible error codes: BAD_REQUEST, VALIDATION_ERROR"
// @Failure     500   {object} common.ErrorResponseDoc "Internal Server Error - Error code: INTERNAL_ERROR"
// @Router      /api/v1/users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.PagingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.validator.ValidatePaging(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	result, err := h.service.GetAll(ctx, &req)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	// Use pagination helper for cleaner response
	common.RespondSuccessWithPagination(c, result.Data, result.Page, result.Limit, result.Total)
}

// @Summary     Update user
// @Description Update an existing user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id   path     string true "User ID"
// @Param       user body     dto.UpdateUserRequest true "User data"
// @Success     200  {object} common.SimpleSuccessResponseDoc
// @Failure     400  {object} common.ErrorResponseDoc "Bad Request - Possible error codes: BAD_REQUEST, VALIDATION_ERROR, EMAIL_EXISTS"
// @Failure     404  {object} common.ErrorResponseDoc "Not Found - Error code: USER_NOT_FOUND"
// @Failure     500  {object} common.ErrorResponseDoc "Internal Server Error - Error code: INTERNAL_ERROR"
// @Router      /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		common.RespondBadRequest(c, "ID is required")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.validator.ValidateUpdate(&req); err != nil {
		common.RespondBadRequest(c, err.Error())
		return
	}

	err := h.service.Update(ctx, id, &req)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccess(c, nil)
}

// @Summary     Delete user
// @Description Delete a user by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id   path     string true "User ID"
// @Success     200  {object} common.SimpleSuccessResponseDoc
// @Failure     404  {object} common.ErrorResponseDoc "Not Found - Error code: USER_NOT_FOUND"
// @Failure     500  {object} common.ErrorResponseDoc "Internal Server Error - Error code: INTERNAL_ERROR"
// @Router      /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		common.RespondBadRequest(c, "ID is required")
		return
	}

	err := h.service.Delete(ctx, id)
	if err != nil {
		common.RespondServiceError(c, err)
		return
	}

	common.RespondSuccess(c, nil)
}
