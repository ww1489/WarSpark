package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	appjwt "github.com/ww1489/WarSpark/pkg/jwt"
)

func TestRequiredRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/private", Required(testTokenManager()), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestRequiredSetsUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := testTokenManager()
	token, _, err := manager.GenerateAccessToken(42, 7)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	router := gin.New()
	router.GET("/private", Required(manager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":   c.GetInt64(UserIDKey),
			"user_role": c.GetInt(UserRoleKey),
		})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if body := recorder.Body.String(); body != `{"user_id":42,"user_role":7}` {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestOptionalIgnoresInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/public", Optional(testTokenManager()), func(c *gin.Context) {
		_, hasUserID := c.Get(UserIDKey)
		c.JSON(http.StatusOK, gin.H{"has_user_id": hasUserID})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/public", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if body := recorder.Body.String(); body != `{"has_user_id":false}` {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func testTokenManager() *appjwt.Manager {
	manager := appjwt.New(appjwt.Config{
		Secret:        "middleware-test-secret",
		Issuer:        "warspark-test",
		AccessExpire:  time.Minute,
		RefreshExpire: time.Hour,
	})
	return manager
}
