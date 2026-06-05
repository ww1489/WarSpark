package utils

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	ErrUnauthorized ErrorCode = 10001
	ErrTokenExpired ErrorCode = 10002
	ErrForbidden    ErrorCode = 10003
	ErrMissingField ErrorCode = 10004
	ErrInvalidField ErrorCode = 10005
	ErrInternal     ErrorCode = 10006
)

type AppError struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	Err        error
}

func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%d: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e AppError) Unwrap() error {
	return e.Err
}

func NewError(code ErrorCode, message string) AppError {
	return AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: StatusForCode(code),
	}
}

func WrapError(code ErrorCode, message string, err error) AppError {
	return AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: StatusForCode(code),
		Err:        err,
	}
}

func ErrorResponse(err error) (int, int, string) {
	var appErr AppError
	if errors.As(err, &appErr) {
		status := appErr.HTTPStatus
		if status == 0 {
			status = StatusForCode(appErr.Code)
		}
		return status, int(appErr.Code), appErr.Message
	}
	return http.StatusInternalServerError, int(ErrInternal), "internal server error"
}

func StatusForCode(code ErrorCode) int {
	switch code {
	case ErrUnauthorized, ErrTokenExpired:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrMissingField, ErrInvalidField:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
