package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"krishblog/pkg/response"
)

// ─── Security Headers ─────────────────────────────────────────────────────────

func SecureHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("Content-Security-Policy",
				"default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'")
			return next(c)
		}
	}
}

// ─── CORS ─────────────────────────────────────────────────────────────────────

func CORS(allowedOrigins []string) echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

// ─── Structured Request Logger ────────────────────────────────────────────────

func Logger(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			req := c.Request()
			res := c.Response()
			log.Info("request",
				slog.String("id", GetRequestID(c)),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("query", req.URL.RawQuery),
				slog.Int("status", res.Status),
				slog.Int64("bytes", res.Size),
				slog.String("ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
				slog.Duration("latency", time.Since(start)),
			)
			return nil
		}
	}
}

// ─── Recover ─────────────────────────────────────────────────────────────────

func Recover(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Error("panic recovered",
						slog.String("id", GetRequestID(c)),
						slog.Any("panic", r),
					)
					_ = response.InternalServerError(c, GetRequestID(c))
				}
			}()
			return next(c)
		}
	}
}

// ─── RBAC ─────────────────────────────────────────────────────────────────────

var roleRank = map[string]int{
	"viewer":     1,
	"editor":     2,
	"admin":      3,
	"superadmin": 4,
}

func RequireRole(minRole string) echo.MiddlewareFunc {
	required := roleRank[strings.ToLower(minRole)]
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaims(c)
			if claims == nil {
				return response.Unauthorized(c, "authentication required")
			}
			if roleRank[strings.ToLower(claims.Role)] < required {
				return response.Forbidden(c, "insufficient permissions")
			}
			return next(c)
		}
	}
}
