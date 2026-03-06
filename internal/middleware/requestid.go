package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get(RequestIDHeader)
			if id == "" {
				id = uuid.NewString()
			}
			c.Set("request_id", id)
			c.Response().Header().Set(RequestIDHeader, id)
			return next(c)
		}
	}
}

func GetRequestID(c echo.Context) string {
	if id, ok := c.Get("request_id").(string); ok {
		return id
	}
	return ""
}
