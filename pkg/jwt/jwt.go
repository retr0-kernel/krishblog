package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	gojwt.RegisteredClaims
}

type Manager struct {
	secret        []byte
	expiry        time.Duration
	refreshExpiry time.Duration
}

func NewManager(secret string, expiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		secret:        []byte(secret),
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
	}
}

func (m *Manager) IssueAccessToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: gojwt.RegisteredClaims{
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(m.expiry)),
			Issuer:    "publishing-platform",
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) IssueRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := gojwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  gojwt.NewNumericDate(now),
		ExpiresAt: gojwt.NewNumericDate(now.Add(m.refreshExpiry)),
		Issuer:    "publishing-platform",
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(
		tokenStr, &Claims{},
		func(t *gojwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		gojwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func (m *Manager) ParseRefreshToken(tokenStr string) (string, error) {
	token, err := gojwt.ParseWithClaims(
		tokenStr, &gojwt.RegisteredClaims{},
		func(t *gojwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}
	claims, ok := token.Claims.(*gojwt.RegisteredClaims)
	if !ok || !token.Valid || claims.Subject == "" {
		return "", errors.New("invalid refresh token claims")
	}
	return claims.Subject, nil
}

func (m *Manager) RefreshExpiry() time.Duration {
	return m.refreshExpiry
}
