package sections

import (
	"errors"

	"github.com/labstack/echo/v4"

	mw "krishblog/internal/middleware"
	"krishblog/pkg/response"
)

// Handler handles section HTTP routes.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListPublic handles GET /v1/public/sections
func (h *Handler) ListPublic(c echo.Context) error {
	secs, err := h.svc.ListActive(c.Request().Context())
	if err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, secs)
}

// GetBySlug handles GET /v1/public/sections/:slug
func (h *Handler) GetBySlug(c echo.Context) error {
	sec, err := h.svc.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFound(c, "section")
		}
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, sec)
}

// AdminList handles GET /v1/admin/sections
func (h *Handler) AdminList(c echo.Context) error {
	secs, err := h.svc.ListAll(c.Request().Context())
	if err != nil {
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, secs)
}

// Create handles POST /v1/admin/sections
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
		if errors.Is(err, ErrSlugTaken) {
			return response.Conflict(c, "slug already in use")
		}
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.Created(c, sec)
}

// Update handles PUT /v1/admin/sections/:id
func (h *Handler) Update(c echo.Context) error {
	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	sec, err := h.svc.Update(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFound(c, "section")
		}
		if errors.Is(err, ErrSlugTaken) {
			return response.Conflict(c, "slug already in use")
		}
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.OK(c, sec)
}

// Delete handles DELETE /v1/admin/sections/:id
func (h *Handler) Delete(c echo.Context) error {
	err := h.svc.Delete(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.NotFound(c, "section")
		}
		c.Logger().Error("failed to delete section: ", err)
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return response.NoContent(c)
}
