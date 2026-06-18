package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	appconfig "github.com/ww1489/WarSpark/internal/config"
	appjwt "github.com/ww1489/WarSpark/pkg/jwt"
)

func TestAdminRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router, appconfig.RuntimeConfig{
		Config: appconfig.Config{},
		TokenManager: appjwt.New(appjwt.Config{
			Secret:        "routes-test-secret",
			Issuer:        "warspark-test",
			AccessExpire:  time.Minute,
			RefreshExpire: time.Hour,
		}),
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/layouts", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	}
}
