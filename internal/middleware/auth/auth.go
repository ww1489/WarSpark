package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ww1489/WarSpark/internal/utils"
	appjwt "github.com/ww1489/WarSpark/pkg/jwt"
)

const (
	AccessTokenKey = "access_token"
	TokenClaimsKey = "token_claims"
	UserIDKey      = "user_id"
	UserRoleKey    = "user_role"
)

func Required(tokenManager *appjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			utils.JSON(c, http.StatusUnauthorized, 10001, "unauthorized", nil)
			c.Abort()
			return
		}

		claims, err := tokenManager.ParseAccessToken(token)
		if err != nil {
			utils.JSON(c, http.StatusUnauthorized, 10001, "invalid access token", nil)
			c.Abort()
			return
		}

		c.Set(AccessTokenKey, token)
		c.Set(TokenClaimsKey, claims)
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Next()
	}
}

func Optional(tokenManager *appjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token := bearerToken(c.GetHeader("Authorization")); token != "" {
			if claims, err := tokenManager.ParseAccessToken(token); err == nil {
				c.Set(AccessTokenKey, token)
				c.Set(TokenClaimsKey, claims)
				c.Set(UserIDKey, claims.UserID)
				c.Set(UserRoleKey, claims.Role)
			}
		}
		c.Next()
	}
}

func UserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok
}

func UserRole(c *gin.Context) (int, bool) {
	value, ok := c.Get(UserRoleKey)
	if !ok {
		return 0, false
	}
	role, ok := value.(int)
	return role, ok
}

func Claims(c *gin.Context) (*appjwt.Claims, bool) {
	value, ok := c.Get(TokenClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*appjwt.Claims)
	return claims, ok
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
