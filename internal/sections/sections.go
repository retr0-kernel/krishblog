package sections

import (
	"context"
	"errors"
	"time"

	"github.com/labstack/echo/v4"

	"krishblog/internal/database"
	mw "krishblog/internal/middleware"
	"krishblog/pkg/response"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type SectionLayout string

const (
	LayoutFeed     SectionLayout = "feed"
	LayoutGrid     SectionLayout = "grid"
	LayoutFeatured SectionLayout = "featured"
	LayoutMinimal  SectionLayout = "minimal"
	LayoutMagazine SectionLayout = "magazine"
)

type CreateRequest struct {
	Name        string                 `json:"name"        validate:"required,min=2,max=100"`
	Slug        string                 `json:"slug"        validate:"omitempty,min=2,max=100"`
	Description string                 `json:"description"`
	Layout      SectionLayout          `json:"layout"      validate:"omitempty,oneof=feed grid featured minimal magazine"`
	CoverImage  string                 `json:"cover_image" validate:"omitempty,url"`
	MetaTitle   string                 `json:"meta_title"  validate:"omitempty,max=70"`
	MetaDesc    string                 `json:"meta_desc"   validate:"omitempty,max=160"`
	IsActive    *bool                  `json:"is_active"`
	SortOrder   int                    `json:"sort_order"`
	Settings    map[string]interface{} `json:"settings"`
}

type SectionResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description,omitempty"`
	Layout      SectionLayout          `json:"layout"`
	CoverImage  string                 `json:"cover_image,omitempty"`
	IsActive    bool                   `json:"is_active"`
	SortOrder   int                    `json:"sort_order"`
	Settings    map[string]interface{} `json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

type Service struct {
	db    *database.Postgres
	redis *database.Redis
}

func NewService(db *database.Postgres, redis *database.Redis) *Service {
	return &Service{db: db, redis: redis}
}

func (s *Service) ListActive(_ context.Context) ([]SectionResponse, error) {
	return []SectionResponse{}, nil
}

func (s *Service) ListAll(_ context.Context) ([]SectionResponse, error) {
	return []SectionResponse{}, nil
}

func (s *Service) GetBySlug(_ context.Context, _ string) (*SectionResponse, error) {
	return nil, errors.New("not found")
}

func (s *Service) Create(_ context.Context, _ CreateRequest) (*SectionResponse, error) {
	return nil, errors.New("not implemented: wire Ent in step 2")
}

func (s *Service) Update(_ context.Context, _ string, _ CreateRequest) (*SectionResponse, error) {
	return nil, errors.New("not implemented: wire Ent in step 2")
}

func (s *Service) Delete(_ context.Context, _ string) error {
	return errors.New("not implemented: wire Ent in step 2")
}

// ─── Handler ──────────────────────────────────────────────────────────────────

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListPublic(c echo.Context) error {
	secs, err := h.svc.ListActive(c.Request().Context())
	if err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, secs)
}

func (h *Handler) GetBySlug(c echo.Context) error {
	sec, err := h.svc.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return response.NotFound(c, "section")
	}
	return response.OK(c, sec)
}

func (h *Handler) AdminList(c echo.Context) error {
	secs, err := h.svc.ListAll(c.Request().Context())
	if err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, secs)
}

func (h *Handler) Create(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	sec, err := h.svc.Create(c.Request().Context(), req)
	if err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.Created(c, sec)
}

func (h *Handler) Update(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	sec, err := h.svc.Update(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, sec)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.NoContent(c)
}
