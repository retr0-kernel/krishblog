package auth

import "time"

// ── Requests ──────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ── Responses ─────────────────────────────────────────────────────────────────

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}

type CSRFResponse struct {
	CSRFToken string `json:"csrf_token"`
}

// ── Internal ──────────────────────────────────────────────────────────────────

// googleUserInfo is the payload from Google's userinfo endpoint.
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// auditAction enumerates audit log action strings.
type auditAction string

const (
	actionLogin        auditAction = "auth.login"
	actionLoginGoogle  auditAction = "auth.login.google"
	actionLoginFailed  auditAction = "auth.login.failed"
	actionLogout       auditAction = "auth.logout"
	actionTokenRefresh auditAction = "auth.token.refresh"
)
