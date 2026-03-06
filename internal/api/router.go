package api

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"krishblog/internal/analytics"
	"krishblog/internal/api/health"
	"krishblog/internal/auth"
	"krishblog/internal/database"
	"krishblog/internal/media"
	mw "krishblog/internal/middleware"
	"krishblog/internal/posts"
	"krishblog/internal/sections"
	jwtpkg "krishblog/pkg/jwt"
)

type Handlers struct {
	Health    *health.Handler
	Auth      *auth.Handler
	Posts     *posts.Handler
	Sections  *sections.Handler
	Analytics *analytics.Handler
	Media     *media.Handler
}

type RouterConfig struct {
	AllowedOrigins []string
	RPS            float64
	Burst          int
	JWTManager     *jwtpkg.Manager
	Redis          *database.Redis
	Logger         *slog.Logger
}

func Register(e *echo.Echo, h Handlers, cfg RouterConfig) {
	// ── Global middleware ────────────────────────────────────────────────────
	e.Use(mw.RequestID())
	e.Use(mw.Recover(cfg.Logger))
	e.Use(mw.Logger(cfg.Logger))
	e.Use(mw.SecureHeaders())
	e.Use(mw.CORS(cfg.AllowedOrigins))
	e.Use(echomw.Decompress())

	// ── System ───────────────────────────────────────────────────────────────
	e.GET("/health", h.Health.Live)
	e.GET("/ready", h.Health.Ready)

	v1 := e.Group("/v1")

	// ── Auth ──────────────────────────────────────────────────────────────────
	authGroup := v1.Group("/auth")
	authGroup.Use(mw.StrictRateLimiter(cfg.Redis, 5))
	authGroup.POST("/login", h.Auth.Login)
	authGroup.POST("/refresh", h.Auth.Refresh)
	authGroup.POST("/logout", h.Auth.Logout, mw.Auth(cfg.JWTManager))
	authGroup.GET("/me", h.Auth.Me, mw.Auth(cfg.JWTManager))

	// ── Public ────────────────────────────────────────────────────────────────
	pub := v1.Group("/public")
	pub.Use(mw.RateLimiter(cfg.Redis, cfg.RPS, cfg.Burst))
	pub.GET("/posts", h.Posts.List)
	pub.GET("/posts/:slug", h.Posts.GetBySlug)
	pub.GET("/sections", h.Sections.ListPublic)
	pub.GET("/sections/:slug", h.Sections.GetBySlug)

	// ── Analytics ─────────────────────────────────────────────────────────────
	aGroup := v1.Group("/analytics")
	aGroup.Use(mw.RateLimiter(cfg.Redis, cfg.RPS*2, cfg.Burst*2))
	aGroup.POST("/event", h.Analytics.RecordEvent)
	aGroup.POST("/session/start", h.Analytics.SessionStart)
	aGroup.POST("/session/end", h.Analytics.SessionEnd)

	// ── Admin ─────────────────────────────────────────────────────────────────
	admin := v1.Group("/admin")
	admin.Use(mw.Auth(cfg.JWTManager))
	admin.Use(mw.RateLimiter(cfg.Redis, cfg.RPS, cfg.Burst))

	// Posts — editor+
	ap := admin.Group("/posts")
	ap.Use(mw.RequireRole("editor"))
	ap.GET("", h.Posts.AdminList)
	ap.POST("", h.Posts.Create)
	ap.PUT("/:id", h.Posts.Update)
	ap.PATCH("/:id/status", h.Posts.UpdateStatus)
	ap.DELETE("/:id", h.Posts.Delete, mw.RequireRole("admin"))

	// Sections — admin+
	as := admin.Group("/sections")
	as.Use(mw.RequireRole("admin"))
	as.GET("", h.Sections.AdminList)
	as.POST("", h.Sections.Create)
	as.PUT("/:id", h.Sections.Update)
	as.DELETE("/:id", h.Sections.Delete)

	// Media — editor+
	am := admin.Group("/media")
	am.Use(mw.RequireRole("editor"))
	am.GET("", h.Media.List)
	am.POST("/upload", h.Media.Upload)
	am.PATCH("/:id", h.Media.Update)
	am.DELETE("/:id", h.Media.Delete)

	// Analytics dashboard — viewer+
	aa := admin.Group("/analytics")
	aa.Use(mw.RequireRole("viewer"))
	aa.GET("/overview", h.Analytics.AdminOverview)
	aa.GET("/posts/:id", h.Analytics.AdminPostStats)
	aa.GET("/realtime", h.Analytics.AdminRealtime)
}
