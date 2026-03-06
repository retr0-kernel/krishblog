package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"krishblog/internal/database"
	jwtpkg "krishblog/pkg/jwt"
	"krishblog/pkg/password"
)

// mockUser simulates a DB row. Replaced by a real repository in Step 2.
type mockUser struct {
	ID           string
	Email        string
	PasswordHash string
	FullName     string
	AvatarURL    string
	Role         string
	IsActive     bool
}

type Service struct {
	redis *database.Redis
	jwt   *jwtpkg.Manager
}

func NewService(redis *database.Redis, jwt *jwtpkg.Manager) *Service {
	return &Service{redis: redis, jwt: jwt}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, string, error) {
	user, err := s.findByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}
	if !user.IsActive {
		return nil, "", errors.New("account is disabled")
	}
	if err := password.Verify(user.PasswordHash, req.Password); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	accessToken, err := s.jwt.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("issue access token: %w", err)
	}
	refreshToken, err := s.jwt.IssueRefreshToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("issue refresh token: %w", err)
	}

	resp := &LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.RefreshExpiry().Seconds()),
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarURL,
			Role:      user.Role,
		},
	}
	return resp, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	revoked, err := s.redis.Exists(ctx, revokedKey(refreshToken))
	if err == nil && revoked {
		return "", errors.New("refresh token has been revoked")
	}

	user, err := s.findByID(ctx, userID)
	if err != nil || !user.IsActive {
		return "", errors.New("user not found or inactive")
	}

	return s.jwt.IssueAccessToken(user.ID, user.Email, user.Role)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.redis.Set(ctx, revokedKey(refreshToken), "1", s.jwt.RefreshExpiry())
}

func (s *Service) RefreshCookieTTL() time.Duration {
	return s.jwt.RefreshExpiry()
}

// ─── stubs — replaced by real repo in Step 2 ─────────────────────────────────

func (s *Service) findByEmail(_ context.Context, _ string) (*mockUser, error) {
	return nil, errors.New("user not found")
}

func (s *Service) findByID(_ context.Context, _ string) (*mockUser, error) {
	return nil, errors.New("user not found")
}

func revokedKey(token string) string {
	return "revoked_rt:" + token
}
