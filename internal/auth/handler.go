package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	mw "krishblog/internal/middleware"
	"krishblog/pkg/response"
)

const (
	refreshCookieName = "refresh_token"
	stateCookieName   = "oauth_state"
)

// Handler handles auth HTTP routes.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Login handles POST /v1/auth/login
func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, refreshToken, err := h.svc.Login(
		c.Request().Context(), req,
		c.RealIP(), c.Request().UserAgent(),
	)
	if err != nil {
		c.Logger().Error("login failed: ", err)
		return response.Unauthorized(c, err.Error())
	}

	setRefreshCookie(c, refreshToken, h.svc.RefreshCookieTTL())
	return response.OK(c, resp)
}

// Refresh handles POST /v1/auth/refresh
func (h *Handler) Refresh(c echo.Context) error {
	refreshToken := extractRefreshToken(c)
	if refreshToken == "" {
		return response.Unauthorized(c, "refresh token required")
	}

	access, newRefresh, err := h.svc.Refresh(c.Request().Context(), refreshToken)
	if err != nil {
		clearRefreshCookie(c)
		return response.Unauthorized(c, err.Error())
	}

	// Rotate cookie
	setRefreshCookie(c, newRefresh, h.svc.RefreshCookieTTL())
	return response.OK(c, map[string]string{
		"access_token": access,
		"token_type":   "Bearer",
	})
}

// Logout handles POST /v1/auth/logout
func (h *Handler) Logout(c echo.Context) error {
	token := extractRefreshToken(c)
	_ = h.svc.Logout(
		c.Request().Context(), token,
		c.RealIP(), c.Request().UserAgent(),
	)
	clearRefreshCookie(c)
	return response.NoContent(c)
}

// Me handles GET /v1/auth/me
func (h *Handler) Me(c echo.Context) error {
	claims := mw.GetClaims(c)
	if claims == nil {
		return response.Unauthorized(c, "not authenticated")
	}
	return response.OK(c, map[string]string{
		"id":    claims.UserID,
		"email": claims.Email,
		"role":  claims.Role,
	})
}

// GoogleLogin handles GET /v1/auth/google
func (h *Handler) GoogleLogin(c echo.Context) error {
	url, err := h.svc.GoogleAuthURL(c.Request().Context())
	if err != nil {
		c.Logger().Error("failed to generate Google auth URL: ", err)
		return response.InternalServerError(c, mw.GetRequestID(c))
	}
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles GET /v1/auth/google/callback
func (h *Handler) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" || state == "" {
		return response.BadRequest(c, "MISSING_PARAMS", "code and state are required", nil)
	}

	resp, refreshToken, err := h.svc.GoogleCallback(
		c.Request().Context(), code, state,
		c.RealIP(), c.Request().UserAgent(),
	)
	if err != nil {
		c.Logger().Error("Google callback failed: ", err)
		return response.Unauthorized(c, err.Error())
	}

	setRefreshCookie(c, refreshToken, h.svc.RefreshCookieTTL())
	return response.OK(c, resp)
}

// ── cookie helpers ────────────────────────────────────────────────────────────

func setRefreshCookie(c echo.Context, token string, ttl time.Duration) {
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

func clearRefreshCookie(c echo.Context) {
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

func extractRefreshToken(c echo.Context) string {
	if cookie, err := c.Cookie(refreshCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	var req RefreshRequest
	if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}
	return ""
}
