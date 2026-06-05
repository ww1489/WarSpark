package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidConfig       = errors.New("invalid jwt config")
	ErrUnexpectedTokenType = errors.New("unexpected token type")
)

type Config struct {
	Secret        string
	Issuer        string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

type Manager struct {
	secret        []byte
	issuer        string
	accessExpire  time.Duration
	refreshExpire time.Duration
	now           func() time.Time
}

type Claims struct {
	UserID    int64  `json:"user_id"`
	Role      int    `json:"role"`
	TokenType string `json:"token_type"`
	jwtv5.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	RefreshJTI       string    `json:"refresh_jti"`
	TokenType        string    `json:"token_type"`
}

func New(cfg Config) *Manager {
	issuer := cfg.Issuer
	if issuer == "" {
		issuer = "app"
	}

	return &Manager{
		secret:        []byte(cfg.Secret),
		issuer:        issuer,
		accessExpire:  cfg.AccessExpire,
		refreshExpire: cfg.RefreshExpire,
		now:           time.Now,
	}
}

func (m *Manager) GeneratePair(userID int64, role int) (TokenPair, error) {
	accessToken, accessExpiresAt, _, err := m.generate(userID, role, TokenTypeAccess, m.accessExpire)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, refreshExpiresAt, refreshJTI, err := m.generate(userID, role, TokenTypeRefresh, m.refreshExpire)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		RefreshJTI:       refreshJTI,
		TokenType:        "Bearer",
	}, nil
}

func (m *Manager) GenerateAccessToken(userID int64, role int) (string, time.Time, error) {
	token, expiresAt, _, err := m.generate(userID, role, TokenTypeAccess, m.accessExpire)
	return token, expiresAt, err
}

func (m *Manager) GenerateRefreshToken(userID int64, role int) (string, time.Time, string, error) {
	return m.generate(userID, role, TokenTypeRefresh, m.refreshExpire)
}

func (m *Manager) Refresh(refreshToken string) (TokenPair, error) {
	claims, err := m.ParseRefreshToken(refreshToken)
	if err != nil {
		return TokenPair{}, err
	}
	return m.GeneratePair(claims.UserID, claims.Role)
}

func (m *Manager) ParseAccessToken(tokenString string) (*Claims, error) {
	return m.parseTyped(tokenString, TokenTypeAccess)
}

func (m *Manager) ParseRefreshToken(tokenString string) (*Claims, error) {
	return m.parseTyped(tokenString, TokenTypeRefresh)
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	if len(m.secret) == 0 {
		return nil, fmt.Errorf("%w: secret is required", ErrInvalidConfig)
	}

	claims := &Claims{}
	token, err := jwtv5.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwtv5.Token) (any, error) {
			if token.Method != jwtv5.SigningMethodHS256 {
				return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, token.Header["alg"])
			}
			return m.secret, nil
		},
		jwtv5.WithIssuer(m.issuer),
		jwtv5.WithExpirationRequired(),
		jwtv5.WithTimeFunc(m.now),
		jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID <= 0 {
		return nil, fmt.Errorf("%w: missing user_id", ErrInvalidToken)
	}
	return claims, nil
}

func (m *Manager) parseTyped(tokenString string, expectedType string) (*Claims, error) {
	claims, err := m.Parse(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%w: want %s got %s", ErrUnexpectedTokenType, expectedType, claims.TokenType)
	}
	return claims, nil
}

func (m *Manager) generate(userID int64, role int, tokenType string, ttl time.Duration) (string, time.Time, string, error) {
	if len(m.secret) == 0 {
		return "", time.Time{}, "", fmt.Errorf("%w: secret is required", ErrInvalidConfig)
	}
	if ttl <= 0 {
		return "", time.Time{}, "", fmt.Errorf("%w: token ttl must be positive", ErrInvalidConfig)
	}
	if userID <= 0 {
		return "", time.Time{}, "", fmt.Errorf("%w: user_id must be positive", ErrInvalidToken)
	}

	now := m.now().UTC()
	expiresAt := now.Add(ttl)
	jti := uuid.NewString()

	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        jti,
			Issuer:    m.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, "", err
	}
	return signed, expiresAt, jti, nil
}
