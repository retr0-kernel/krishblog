package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func OKWithMeta(c echo.Context, data interface{}, meta interface{}) error {
	return c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func BadRequest(c echo.Context, code, message string, details interface{}) error {
	return c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Error:   &APIError{Code: code, Message: message, Details: details},
	})
}

func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, Envelope{
		Success: false,
		Error:   &APIError{Code: "UNAUTHORIZED", Message: message},
	})
}

func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, Envelope{
		Success: false,
		Error:   &APIError{Code: "FORBIDDEN", Message: message},
	})
}

func NotFound(c echo.Context, resource string) error {
	return c.JSON(http.StatusNotFound, Envelope{
		Success: false,
		Error:   &APIError{Code: "NOT_FOUND", Message: resource + " not found"},
	})
}

func Conflict(c echo.Context, message string) error {
	return c.JSON(http.StatusConflict, Envelope{
		Success: false,
		Error:   &APIError{Code: "CONFLICT", Message: message},
	})
}

func UnprocessableEntity(c echo.Context, details interface{}) error {
	return c.JSON(http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Error:   &APIError{Code: "VALIDATION_ERROR", Message: "request validation failed", Details: details},
	})
}

func TooManyRequests(c echo.Context) error {
	return c.JSON(http.StatusTooManyRequests, Envelope{
		Success: false,
		Error:   &APIError{Code: "RATE_LIMITED", Message: "too many requests, please slow down"},
	})
}

func InternalServerError(c echo.Context, requestID string) error {
	return c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
			Details: map[string]string{"request_id": requestID},
		},
	})
}
