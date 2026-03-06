package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"krishblog/internal/database"
)

type Dependencies struct {
	Postgres  *database.Postgres
	Redis     *database.Redis
	Version   string
	StartedAt time.Time
}

type Handler struct {
	deps Dependencies
}

func New(deps Dependencies) *Handler {
	return &Handler{deps: deps}
}

type healthStatus struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func (h *Handler) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allOK := true

	if err := h.deps.Postgres.Health(ctx); err != nil {
		services["postgres"] = "unhealthy: " + err.Error()
		allOK = false
	} else {
		services["postgres"] = "ok"
	}

	if err := h.deps.Redis.Health(ctx); err != nil {
		services["redis"] = "unhealthy: " + err.Error()
		allOK = false
	} else {
		services["redis"] = "ok"
	}

	status := "ok"
	code := http.StatusOK
	if !allOK {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	return c.JSON(code, healthStatus{
		Status:    status,
		Version:   h.deps.Version,
		Uptime:    time.Since(h.deps.StartedAt).Round(time.Second).String(),
		Timestamp: time.Now().UTC(),
		Services:  services,
	})
}
