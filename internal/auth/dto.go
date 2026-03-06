package auth

import "time"

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}

// tokenMeta is used internally to track issued refresh tokens.
type tokenMeta struct {
	UserID    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}
