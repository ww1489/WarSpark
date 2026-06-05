package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindJSONReportsValidationFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type request struct {
		Email string `json:"email" binding:"required,email"`
	}

	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		var req request
		if !BindJSON(c, &req) {
			return
		}
		OK(c, nil)
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"email":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"field":"email"`) {
		t.Fatalf("expected json field name in response, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"rule":"email"`) {
		t.Fatalf("expected validation rule in response, got %s", recorder.Body.String())
	}
}
