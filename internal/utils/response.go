package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, 0, "ok", data)
}

func Fail(c *gin.Context, err error) {
	httpStatus, code, message := ErrorResponse(err)
	JSON(c, httpStatus, code, message, nil)
}

func JSON(c *gin.Context, httpStatus int, code int, message string, data any) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
