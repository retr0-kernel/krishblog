package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"krishblog/internal/analytics"
	"krishblog/internal/api"
	"krishblog/internal/api/health"
	"krishblog/internal/auth"
	"krishblog/internal/config"
	"krishblog/internal/database"
	"krishblog/internal/media"
	mw "krishblog/internal/middleware"
	"krishblog/internal/posts"
	"krishblog/internal/sections"
	jwtpkg "krishblog/pkg/jwt"
	"krishblog/pkg/logger"
	"krishblog/pkg/validator"
)

var version = "dev"

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		panic("config load failed: " + err.Error())
	}

	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.New(cfg.App.Env)
	startedAt := time.Now()

	log.Info("starting server",
		"version", version,
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	// ── Postgres ──────────────────────────────────────────────────────────────
	pg, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Error("postgres connect failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := pg.Close(); cerr != nil {
			log.Error("postgres close error", "error", cerr)
		}
	}()
	log.Info("postgres connected")

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb, err := database.NewRedis(cfg.Redis.URL)
	if err != nil {
		log.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := rdb.Close(); cerr != nil {
			log.Error("redis close error", "error", cerr)
		}
	}()
	log.Info("redis connected")

	// ── JWT ───────────────────────────────────────────────────────────────────
	jwtManager := jwtpkg.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.ExpiryHours,
		cfg.JWT.RefreshExpiryHours,
	)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := auth.NewService(rdb, jwtManager)
	postsSvc := posts.NewService(pg, rdb)
	sectionsSvc := sections.NewService(pg, rdb)
	analyticsSvc := analytics.NewService(pg, rdb)
	mediaSvc := media.NewService(pg, rdb)

	// ── Handlers ──────────────────────────────────────────────────────────────
	handlers := api.Handlers{
		Health: health.New(health.Dependencies{
			Postgres:  pg,
			Redis:     rdb,
			Version:   version,
			StartedAt: startedAt,
		}),
		Auth:      auth.NewHandler(authSvc),
		Posts:     posts.NewHandler(postsSvc),
		Sections:  sections.NewHandler(sectionsSvc),
		Analytics: analytics.NewHandler(analyticsSvc),
		Media:     media.NewHandler(mediaSvc),
	}

	// ── Echo ──────────────────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var he *echo.HTTPError
		if errors.As(err, &he) {
			_ = c.JSON(he.Code, map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"code":    http.StatusText(he.Code),
					"message": he.Message,
				},
			})
			return
		}
		_ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": "an unexpected error occurred",
			},
		})
	}

	// ── Routes ────────────────────────────────────────────────────────────────
	api.Register(e, handlers, api.RouterConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		RPS:            cfg.RateLimit.RPS,
		Burst:          cfg.RateLimit.Burst,
		JWTManager:     jwtManager,
		Redis:          rdb,
		Logger:         log,
	})

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	serverErr := make(chan error, 1)
	go func() {
		addr := ":" + cfg.App.Port
		log.Info("listening", "addr", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}

// Ensure mw import is used (referenced in router, not main — kept for clarity).
var _ = mw.GetRequestID
