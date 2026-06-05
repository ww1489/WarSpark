package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestManagerGenerateAndParseAccessToken(t *testing.T) {
	manager := testManager()

	token, expiresAt, err := manager.GenerateAccessToken(42, 9)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if expiresAt.IsZero() {
		t.Fatal("expiresAt should not be zero")
	}

	claims, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("unexpected user id: %d", claims.UserID)
	}
	if claims.Role != 9 {
		t.Fatalf("unexpected role: %d", claims.Role)
	}
	if claims.TokenType != TokenTypeAccess {
		t.Fatalf("unexpected token type: %s", claims.TokenType)
	}
}

func TestManagerRejectsRefreshTokenAsAccessToken(t *testing.T) {
	manager := testManager()

	refreshToken, _, _, err := manager.GenerateRefreshToken(42, 0)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	_, err = manager.ParseAccessToken(refreshToken)
	if !errors.Is(err, ErrUnexpectedTokenType) {
		t.Fatalf("expected ErrUnexpectedTokenType, got %v", err)
	}
}

func TestManagerRefreshCreatesNewPair(t *testing.T) {
	manager := testManager()

	refreshToken, _, _, err := manager.GenerateRefreshToken(42, 0)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	pair, err := manager.Refresh(refreshToken)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("token pair should contain both tokens")
	}
	if pair.RefreshJTI == "" {
		t.Fatal("refresh jti should not be empty")
	}
}

func TestManagerRejectsInvalidGenerateInputs(t *testing.T) {
	manager := testManager()

	_, _, err := manager.GenerateAccessToken(0, 0)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for invalid user id, got %v", err)
	}

	manager.accessExpire = 0
	_, _, err = manager.GenerateAccessToken(42, 0)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig for invalid ttl, got %v", err)
	}

	manager = testManager()
	manager.secret = nil
	_, _, err = manager.GenerateAccessToken(42, 0)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig for missing secret, got %v", err)
	}
}

func testManager() *Manager {
	manager := New(Config{
		Secret:        "unit-test-secret",
		Issuer:        "warspark-test",
		AccessExpire:  time.Minute,
		RefreshExpire: time.Hour,
	})
	manager.now = func() time.Time {
		return time.Unix(1_700_000_000, 0)
	}
	return manager
}
