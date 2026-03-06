package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	mw "krishblog/internal/middleware"
	"krishblog/pkg/response"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	loginResp, refreshToken, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	h.setRefreshCookie(c, refreshToken, h.svc.RefreshCookieTTL())
	return response.OK(c, loginResp)
}

func (h *Handler) Refresh(c echo.Context) error {
	refreshToken := ""

	if cookie, err := c.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	}
	if refreshToken == "" {
		var req RefreshRequest
		if err := c.Bind(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}
	if refreshToken == "" {
		return response.Unauthorized(c, "refresh token is required")
	}

	accessToken, err := h.svc.Refresh(c.Request().Context(), refreshToken)
	if err != nil {
		h.clearRefreshCookie(c)
		return response.Unauthorized(c, err.Error())
	}

	return response.OK(c, map[string]string{
		"access_token": accessToken,
		"token_type":   "Bearer",
	})
}

func (h *Handler) Logout(c echo.Context) error {
	if cookie, err := c.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		_ = h.svc.Logout(c.Request().Context(), cookie.Value)
	}
	h.clearRefreshCookie(c)
	return response.NoContent(c)
}

func (h *Handler) Me(c echo.Context) error {
	claims := mw.GetClaims(c)
	if claims == nil {
		return response.Unauthorized(c, "not authenticated")
	}
	return response.OK(c, UserResponse{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	})
}

func (h *Handler) setRefreshCookie(c echo.Context, token string, ttl time.Duration) {
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

func (h *Handler) clearRefreshCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
