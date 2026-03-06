package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"krishblog/internal/database"
	//mw "krishblog/internal/middleware"
	"krishblog/pkg/response"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type EventType string

const (
	EventPageView     EventType = "page_view"
	EventPostView     EventType = "post_view"
	EventScrollDepth  EventType = "scroll_depth"
	EventClick        EventType = "click"
	EventSearch       EventType = "search"
	EventSessionStart EventType = "session_start"
	EventSessionEnd   EventType = "session_end"
)

type EventRequest struct {
	Type      EventType              `json:"type"       validate:"required,oneof=page_view post_view scroll_depth click search session_start session_end"`
	SessionID string                 `json:"session_id" validate:"required,min=10,max=100"`
	PostID    string                 `json:"post_id"    validate:"omitempty,uuid4"`
	SectionID string                 `json:"section_id" validate:"omitempty,uuid4"`
	Path      string                 `json:"path"       validate:"required"`
	Referrer  string                 `json:"referrer"`
	ScrollPct *int                   `json:"scroll_pct" validate:"omitempty,min=0,max=100"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type SessionStartRequest struct {
	SessionID string `json:"session_id" validate:"required,min=10,max=100"`
	Referrer  string `json:"referrer"`
	Path      string `json:"path"       validate:"required"`
}

type SessionEndRequest struct {
	SessionID  string `json:"session_id"  validate:"required,min=10,max=100"`
	DurationMs int    `json:"duration_ms" validate:"min=0"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

const analyticsBufferKey = "analytics:buffer"

type Service struct {
	redis *database.Redis
	db    *database.Postgres
}

func NewService(db *database.Postgres, redis *database.Redis) *Service {
	return &Service{db: db, redis: redis}
}

func (s *Service) RecordEvent(ctx context.Context, req EventRequest, ip, userAgent, country string) error {
	event := map[string]interface{}{
		"type":        req.Type,
		"session_id":  req.SessionID,
		"post_id":     req.PostID,
		"section_id":  req.SectionID,
		"path":        req.Path,
		"referrer":    req.Referrer,
		"scroll_pct":  req.ScrollPct,
		"metadata":    req.Metadata,
		"ip":          ip,
		"user_agent":  userAgent,
		"country":     country,
		"recorded_at": time.Now().UTC().Format(time.RFC3339),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal analytics event: %w", err)
	}
	return s.redis.RPush(ctx, analyticsBufferKey, payload)
}

func (s *Service) RecordSessionStart(ctx context.Context, req SessionStartRequest, ip, userAgent string) error {
	return s.RecordEvent(ctx, EventRequest{
		Type:      EventSessionStart,
		SessionID: req.SessionID,
		Path:      req.Path,
		Referrer:  req.Referrer,
	}, ip, userAgent, "")
}

func (s *Service) RecordSessionEnd(ctx context.Context, req SessionEndRequest) error {
	return s.RecordEvent(ctx, EventRequest{
		Type:      EventSessionEnd,
		SessionID: req.SessionID,
		Path:      "/",
		Metadata:  map[string]interface{}{"duration_ms": req.DurationMs},
	}, "", "", "")
}

// ─── Handler ──────────────────────────────────────────────────────────────────

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RecordEvent(c echo.Context) error {
	var req EventRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	// Fail silently — analytics must never block the user.
	_ = h.svc.RecordEvent(c.Request().Context(), req,
		c.RealIP(),
		c.Request().UserAgent(),
		c.Request().Header.Get("CF-IPCountry"),
	)
	return response.NoContent(c)
}

func (h *Handler) SessionStart(c echo.Context) error {
	var req SessionStartRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	_ = h.svc.RecordSessionStart(c.Request().Context(), req, c.RealIP(), c.Request().UserAgent())
	return response.NoContent(c)
}

func (h *Handler) SessionEnd(c echo.Context) error {
	var req SessionEndRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	_ = h.svc.RecordSessionEnd(c.Request().Context(), req)
	return response.NoContent(c)
}

func (h *Handler) AdminOverview(c echo.Context) error {
	return response.OK(c, map[string]interface{}{
		"total_views":  0,
		"sessions":     0,
		"avg_scroll":   0,
		"top_posts":    []interface{}{},
		"top_sections": []interface{}{},
	})
}

func (h *Handler) AdminPostStats(c echo.Context) error {
	return response.OK(c, map[string]interface{}{
		"post_id": c.Param("id"),
		"daily":   []interface{}{},
	})
}

func (h *Handler) AdminRealtime(c echo.Context) error {
	return response.OK(c, map[string]interface{}{
		"active_sessions": 0,
		"recent_events":   []interface{}{},
	})
}
