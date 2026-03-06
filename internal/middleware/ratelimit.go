package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"krishblog/internal/database"
	"krishblog/pkg/response"
)

func RateLimiter(rdb *database.Redis, rps float64, burst int) echo.MiddlewareFunc {
	windowDur := time.Second

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := fmt.Sprintf("rl:%s", c.RealIP())
			ctx := context.Background()

			count, err := rdb.IncrWithTTL(ctx, key, windowDur)
			if err != nil {
				// Fail open: if Redis is down, allow the request.
				return next(c)
			}

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", burst))
			remaining := int64(burst) - count
			if remaining < 0 {
				remaining = 0
			}
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if count > int64(burst) {
				c.Response().Header().Set("Retry-After", "1")
				return response.TooManyRequests(c)
			}
			return next(c)
		}
	}
}

func StrictRateLimiter(rdb *database.Redis, burst int) echo.MiddlewareFunc {
	return RateLimiter(rdb, float64(burst)/60.0, burst)
}
