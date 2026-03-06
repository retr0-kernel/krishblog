package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	JWT           JWTConfig
	R2            R2Config
	CORS          CORSConfig
	RateLimit     RateLimitConfig
	Admin         AdminConfig
	Google        GoogleConfig
	AllowedAdmins []string
}

type AppConfig struct {
	Env    string
	Port   string
	Secret string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret             string
	ExpiryHours        time.Duration
	RefreshExpiryHours time.Duration
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

type AdminConfig struct {
	Email    string
	Password string
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func Load() (*Config, error) {
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	cfg := &Config{}

	cfg.App.Env = getRequired("APP_ENV")
	cfg.App.Port = getDefault("APP_PORT", getDefault("PORT", "8080"))
	cfg.App.Secret = getRequired("APP_SECRET")

	cfg.Database.URL = getRequired("DATABASE_URL")
	cfg.Redis.URL = getRequired("REDIS_URL")

	cfg.JWT.Secret = getRequired("JWT_SECRET")

	jwtH, err := strconv.Atoi(getDefault("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}
	cfg.JWT.ExpiryHours = time.Duration(jwtH) * time.Hour

	refreshH, err := strconv.Atoi(getDefault("JWT_REFRESH_EXPIRY_HOURS", "168"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY_HOURS: %w", err)
	}
	cfg.JWT.RefreshExpiryHours = time.Duration(refreshH) * time.Hour

	//cfg.R2.AccountID = getRequired("R2_ACCOUNT_ID")
	//cfg.R2.AccessKeyID = getRequired("R2_ACCESS_KEY_ID")
	//cfg.R2.SecretAccessKey = getRequired("R2_SECRET_ACCESS_KEY")
	//cfg.R2.BucketName = getRequired("R2_BUCKET_NAME")
	//cfg.R2.PublicURL = getRequired("R2_PUBLIC_URL")

	cfg.R2.AccountID = getDefault("R2_ACCOUNT_ID", "")
	cfg.R2.AccessKeyID = getDefault("R2_ACCESS_KEY_ID", "")
	cfg.R2.SecretAccessKey = getDefault("R2_SECRET_ACCESS_KEY", "")
	cfg.R2.BucketName = getDefault("R2_BUCKET_NAME", "")
	cfg.R2.PublicURL = getDefault("R2_PUBLIC_URL", "")

	originsRaw := getDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.CORS.AllowedOrigins = strings.Split(originsRaw, ",")

	rps, err := strconv.ParseFloat(getDefault("RATE_LIMIT_RPS", "20"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
	}
	cfg.RateLimit.RPS = rps

	burst, err := strconv.Atoi(getDefault("RATE_LIMIT_BURST", "50"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}
	cfg.RateLimit.Burst = burst

	cfg.Admin.Email = getDefault("ADMIN_EMAIL", "")
	cfg.Admin.Password = getDefault("ADMIN_PASSWORD", "")

	// Google OAuth
	cfg.Google.ClientID = getDefault("GOOGLE_CLIENT_ID", "")
	cfg.Google.ClientSecret = getDefault("GOOGLE_CLIENT_SECRET", "")
	cfg.Google.RedirectURL = getDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/v1/auth/google/callback")

	// Allowed admin emails (comma-separated)
	allowedRaw := getDefault("ADMIN_ALLOWED_EMAILS", cfg.Admin.Email)
	if allowedRaw != "" {
		for _, e := range strings.Split(allowedRaw, ",") {
			trimmed := strings.TrimSpace(strings.ToLower(e))
			if trimmed != "" {
				cfg.AllowedAdmins = append(cfg.AllowedAdmins, trimmed)
			}
		}
	}

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Env != "production"
}

func (c *Config) IsAdminAllowed(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	for _, allowed := range c.AllowedAdmins {
		if allowed == email {
			return true
		}
	}
	return false
}

func getRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("[config] required env var %q is not set", key))
	}
	return v
}

func getDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
