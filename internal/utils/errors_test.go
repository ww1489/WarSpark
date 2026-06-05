package utils

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorResponseUsesAppError(t *testing.T) {
	status, code, message := ErrorResponse(NewError(ErrForbidden, "forbidden"))
	if status != http.StatusForbidden {
		t.Fatalf("unexpected status: %d", status)
	}
	if code != int(ErrForbidden) {
		t.Fatalf("unexpected code: %d", code)
	}
	if message != "forbidden" {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestErrorResponseFallsBackToInternal(t *testing.T) {
	status, code, message := ErrorResponse(errors.New("boom"))
	if status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", status)
	}
	if code != int(ErrInternal) {
		t.Fatalf("unexpected code: %d", code)
	}
	if message != "internal server error" {
		t.Fatalf("unexpected message: %s", message)
	}
}
