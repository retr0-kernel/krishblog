package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"krishblog/ent"
	"krishblog/ent/user"
	"krishblog/internal/config"
	"krishblog/internal/database"
	jwtpkg "krishblog/pkg/jwt"
	"krishblog/pkg/password"
	"krishblog/pkg/uuidutil"
)

// Service handles all authentication business logic.
type Service struct {
	ent         *ent.Client
	redis       *database.Redis
	jwt         *jwtpkg.Manager
	cfg         *config.Config
	log         *slog.Logger
	oauthConfig *oauth2.Config
}

func NewService(
	entClient *ent.Client,
	redis *database.Redis,
	jwt *jwtpkg.Manager,
	cfg *config.Config,
	log *slog.Logger,
) *Service {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return &Service{
		ent:         entClient,
		redis:       redis,
		jwt:         jwt,
		cfg:         cfg,
		log:         log,
		oauthConfig: oauthConfig,
	}
}

// Login validates email+password credentials.
func (s *Service) Login(ctx context.Context, req LoginRequest, ip, ua string) (*LoginResponse, string, error) {
	u, err := s.ent.User.
		Query().
		Where(user.EmailEQ(strings.ToLower(req.Email))).
		Only(ctx)
	if err != nil {
		s.audit(ctx, nil, actionLoginFailed, ip, ua, "user not found")
		_, _ = password.Hash("timing-prevention-dummy")
		return nil, "", errors.New("invalid credentials")
	}

	if !u.IsActive {
		s.audit(ctx, &u.ID, actionLoginFailed, ip, ua, "account inactive")
		return nil, "", errors.New("account is disabled")
	}

	if !s.cfg.IsAdminAllowed(u.Email) {
		s.audit(ctx, &u.ID, actionLoginFailed, ip, ua, "not in allowlist")
		return nil, "", errors.New("access denied")
	}

	if err := password.Verify(u.PasswordHash, req.Password); err != nil {
		s.audit(ctx, &u.ID, actionLoginFailed, ip, ua, "wrong password")
		return nil, "", errors.New("invalid credentials")
	}

	now := time.Now()
	_, _ = s.ent.User.UpdateOneID(u.ID).SetLastLoginAt(now).Save(ctx)

	return s.issueTokens(ctx, u, actionLogin, ip, ua)
}

// GoogleAuthURL returns the consent URL and state token.
func (s *Service) GoogleAuthURL(ctx context.Context) (string, error) {
	state, err := secureToken(32)
	if err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	if err := s.redis.Set(ctx, oauthStateKey(state), "1", 10*time.Minute); err != nil {
		return "", fmt.Errorf("store state: %w", err)
	}
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// GoogleCallback handles the OAuth2 callback.
func (s *Service) GoogleCallback(ctx context.Context, code, state, ip, ua string) (*LoginResponse, string, error) {
	exists, err := s.redis.Exists(ctx, oauthStateKey(state))
	if err != nil || !exists {
		return nil, "", errors.New("invalid or expired oauth state")
	}
	_ = s.redis.Delete(ctx, oauthStateKey(state))

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("oauth code exchange: %w", err)
	}

	info, err := s.googleUserInfo(ctx, token)
	if err != nil {
		return nil, "", err
	}

	if !info.VerifiedEmail {
		return nil, "", errors.New("google email not verified")
	}

	email := strings.ToLower(info.Email)
	if !s.cfg.IsAdminAllowed(email) {
		return nil, "", errors.New("email not authorised for admin access")
	}

	u, err := s.upsertGoogleUser(ctx, info)
	if err != nil {
		return nil, "", err
	}

	return s.issueTokens(ctx, u, actionLoginGoogle, ip, ua)
}

// Refresh rotates the refresh token and issues a new access token.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if revoked, _ := s.redis.Exists(ctx, revokedKey(refreshToken)); revoked {
		return "", "", errors.New("refresh token revoked")
	}

	uid, err := uuidutil.Parse(userID)
	if err != nil {
		return "", "", errors.New("malformed token subject")
	}

	u, err := s.ent.User.Get(ctx, uid)
	if err != nil || !u.IsActive {
		return "", "", errors.New("user not found or inactive")
	}

	// Rotate: revoke old token
	_ = s.redis.Set(ctx, revokedKey(refreshToken), "1", s.jwt.RefreshExpiry())

	access, err := s.jwt.IssueAccessToken(u.ID.String(), u.Email, string(u.Role))
	if err != nil {
		return "", "", fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.jwt.IssueRefreshToken(u.ID.String())
	if err != nil {
		return "", "", fmt.Errorf("issue refresh token: %w", err)
	}

	s.audit(ctx, &u.ID, actionTokenRefresh, "", "", "")
	return access, refresh, nil
}

// Logout revokes the refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken, ip, ua string) error {
	if refreshToken == "" {
		return nil
	}
	if userID, err := s.jwt.ParseRefreshToken(refreshToken); err == nil {
		if uid, err := uuidutil.Parse(userID); err == nil {
			s.audit(ctx, &uid, actionLogout, ip, ua, "")
		}
	}
	return s.redis.Set(ctx, revokedKey(refreshToken), "1", s.jwt.RefreshExpiry())
}

func (s *Service) RefreshCookieTTL() time.Duration { return s.jwt.RefreshExpiry() }

// ── internal ──────────────────────────────────────────────────────────────────

func (s *Service) issueTokens(ctx context.Context, u *ent.User, action auditAction, ip, ua string) (*LoginResponse, string, error) {
	access, err := s.jwt.IssueAccessToken(u.ID.String(), u.Email, string(u.Role))
	if err != nil {
		return nil, "", fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.jwt.IssueRefreshToken(u.ID.String())
	if err != nil {
		return nil, "", fmt.Errorf("issue refresh token: %w", err)
	}
	s.audit(ctx, &u.ID, action, ip, ua, "")
	return &LoginResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.RefreshExpiry().Seconds()),
		User:        toUserResponse(u),
	}, refresh, nil
}

func (s *Service) googleUserInfo(ctx context.Context, token *oauth2.Token) (*googleUserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo: %w", err)
	}
	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse userinfo: %w", err)
	}
	return &info, nil
}

func (s *Service) upsertGoogleUser(ctx context.Context, info *googleUserInfo) (*ent.User, error) {
	email := strings.ToLower(info.Email)

	u, err := s.ent.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err == nil {
		return s.ent.User.UpdateOneID(u.ID).
			SetAvatarURL(info.Picture).
			SetLastLoginAt(time.Now()).
			Save(ctx)
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("query user: %w", err)
	}

	dummy, _ := password.Hash(generateDummy())
	return s.ent.User.Create().
		SetEmail(email).
		SetPasswordHash(dummy).
		SetFullName(info.Name).
		SetAvatarURL(info.Picture).
		SetRole(user.RoleEditor).
		SetIsActive(true).
		Save(ctx)
}

func (s *Service) audit(ctx context.Context, uid *uuid.UUID, action auditAction, ip, ua, note string) {
	uidStr := ""
	if uid != nil {
		uidStr = uid.String()
	}
	s.log.Info("audit",
		slog.String("action", string(action)),
		slog.String("user_id", uidStr),
		slog.String("ip", ip),
		slog.String("ua", ua),
		slog.String("note", note),
	)
}

func revokedKey(t string) string    { return "revoked_rt:" + t }
func oauthStateKey(s string) string { return "oauth_state:" + s }

func secureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateDummy() string { t, _ := secureToken(32); return t }

func toUserResponse(u *ent.User) UserResponse {
	r := UserResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
	r.AvatarURL = u.AvatarURL
	return r
}
