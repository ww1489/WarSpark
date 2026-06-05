package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewUsesExistingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(New())
	router.GET("/", func(c *gin.Context) {
		value, _ := c.Get(Key)
		c.JSON(http.StatusOK, gin.H{"request_id": value})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "request-123")
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Request-ID") != "request-123" {
		t.Fatalf("unexpected response request id: %s", recorder.Header().Get("X-Request-ID"))
	}
	if body := recorder.Body.String(); body != `{"request_id":"request-123"}` {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestNewGeneratesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(New())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected generated request id header")
	}
}
