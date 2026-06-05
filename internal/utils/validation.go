package utils

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type FieldViolation struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
	Param string `json:"param,omitempty"`
}

func init() {
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return field.Name
			}
			return name
		})
	}
}

func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		ValidationFailed(c, err)
		return false
	}
	return true
}

func BindQuery(c *gin.Context, dst any) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		ValidationFailed(c, err)
		return false
	}
	return true
}

func Validate(c *gin.Context, value any) bool {
	if err := binding.Validator.ValidateStruct(value); err != nil {
		ValidationFailed(c, err)
		return false
	}
	return true
}

func ValidationFailed(c *gin.Context, err error) {
	violations := fieldViolations(err)
	if len(violations) == 0 {
		JSON(c, http.StatusBadRequest, int(ErrInvalidField), "invalid request", nil)
		return
	}
	JSON(c, http.StatusUnprocessableEntity, int(ErrInvalidField), "validation failed", gin.H{
		"fields": violations,
	})
}

func fieldViolations(err error) []FieldViolation {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	violations := make([]FieldViolation, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		violations = append(violations, FieldViolation{
			Field: fieldErr.Field(),
			Rule:  fieldErr.Tag(),
			Param: fieldErr.Param(),
		})
	}
	return violations
}
