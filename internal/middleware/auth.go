package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	jwtpkg "krishblog/pkg/jwt"
	"krishblog/pkg/response"
)

const ClaimsKey = "jwt_claims"

func Auth(manager *jwtpkg.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := extractBearer(c)
			if err != nil {
				return response.Unauthorized(c, "missing or malformed authorization header")
			}
			claims, err := manager.ParseAccessToken(token)
			if err != nil {
				return response.Unauthorized(c, "invalid or expired token")
			}
			c.Set(ClaimsKey, claims)
			return next(c)
		}
	}
}

func OptionalAuth(manager *jwtpkg.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if token, err := extractBearer(c); err == nil {
				if claims, err := manager.ParseAccessToken(token); err == nil {
					c.Set(ClaimsKey, claims)
				}
			}
			return next(c)
		}
	}
}

func GetClaims(c echo.Context) *jwtpkg.Claims {
	if claims, ok := c.Get(ClaimsKey).(*jwtpkg.Claims); ok {
		return claims
	}
	return nil
}

func extractBearer(c echo.Context) (string, error) {
	header := c.Request().Header.Get("Authorization")
	if header == "" {
		return "", echo.ErrUnauthorized
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return "", echo.ErrUnauthorized
	}
	return parts[1], nil
}
