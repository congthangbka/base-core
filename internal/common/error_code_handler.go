package common

import "github.com/gin-gonic/gin"

// GetErrorCodes handles the error codes endpoint
// @Summary     Get all error codes
// @Description Get a complete list of all error codes with their messages and HTTP status codes
// @Tags        common
// @Accept      json
// @Produce     json
// @Success     200  {object} SuccessResponseDoc{data=ErrorCodesResponse}
// @Router      /error-codes [get]
func GetErrorCodes(c *gin.Context) {
	errorCodes := GetAllErrorCodes()
	RespondSuccess(c, errorCodes)
}
