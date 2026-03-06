package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	_ "strings"

	"krishblog/pkg/response"

	"github.com/labstack/echo/v4"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenLen   = 32
)

// CSRF implements the double-submit cookie pattern.
// On every request:
//   - If no CSRF cookie exists, generate and set one.
//   - On state-mutating methods (POST/PUT/PATCH/DELETE), verify that the
//     X-CSRF-Token header matches the cookie value.
func CSRF() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Generate token if not present
			token, err := getOrCreateCSRFToken(c)
			if err != nil {
				return response.InternalServerError(c, GetRequestID(c))
			}

			// Expose token to the frontend via a readable header
			c.Response().Header().Set(csrfHeaderName, token)

			// Validate on state-mutating methods
			method := c.Request().Method
			if method == http.MethodPost ||
				method == http.MethodPut ||
				method == http.MethodPatch ||
				method == http.MethodDelete {

				headerToken := c.Request().Header.Get(csrfHeaderName)
				if !secureCompare(token, headerToken) {
					return c.JSON(http.StatusForbidden, map[string]interface{}{
						"success": false,
						"error": map[string]string{
							"code":    "CSRF_INVALID",
							"message": "invalid or missing CSRF token",
						},
					})
				}
			}

			return next(c)
		}
	}
}

func getOrCreateCSRFToken(c echo.Context) (string, error) {
	if cookie, err := c.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	b := make([]byte, csrfTokenLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	c.SetCookie(&http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Must be readable by JS to send in header
		Secure:   c.Request().TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	return token, nil
}

// secureCompare does a constant-time string comparison to prevent timing attacks.
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// SkipCSRF is a helper to mark routes that intentionally skip CSRF
// (e.g. mobile API clients that use Bearer tokens instead of cookies).
func SkipCSRF(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// CSRFToken returns the current CSRF token for inclusion in API responses.
func CSRFToken(c echo.Context) string {
	if cookie, err := c.Cookie(csrfCookieName); err == nil {
		return cookie.Value
	}
	return ""
}
