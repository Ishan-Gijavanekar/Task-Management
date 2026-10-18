package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/config"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	httpmiddleware "github.com/Ishan-Gijavanekar/user-service/internal/transport/http/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *slog.Logger
}

func (s *Server) Validate() error {
	if s == nil || s.app == nil {
		return fmt.Errorf("http server: app is required")
	}
	if s.logger == nil {
		return fmt.Errorf("http server: logger is required")
	}
	return nil
}

func NewServer(
	cfg config.HTTPConfig,
	authCfg config.AuthConfig,
	userService *service.UserService,
	authService *service.AuthService,
	jwtManager *auth.JWTManager,
	logger *slog.Logger,
	readinessCheckers ...ReadinessChecker,
) *Server {
	var readinessChecker ReadinessChecker
	if len(readinessCheckers) > 0 {
		readinessChecker = readinessCheckers[0]
	}
	app := fiber.New(fiber.Config{
		AppName:                 "user-service",
		DisableStartupMessage:   true,
		BodyLimit:               cfg.BodyLimitMB * 1024 * 1024,
		ErrorHandler:            NewErrorHandler(logger),
		ProxyHeader:             fiber.HeaderXForwardedFor,
		EnableTrustedProxyCheck: len(cfg.TrustedProxies) > 0,
		TrustedProxies:          cfg.TrustedProxies,
	})

	app.Use(requestid.New())
	app.Use(httpmiddleware.Logging(logger))
	app.Use(recover.New())
	app.Use(httpmiddleware.Security())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	authHandler := NewAuthHandler(authService, authCfg)
	userHandler := NewUserHandler(userService)
	authMiddleware := httpmiddleware.NewAuthMiddleware(jwtManager)

	registerRoutes(
		app,
		authHandler,
		userHandler,
		authMiddleware,
		readinessChecker,
	)

	return &Server{
		app:    app,
		port:   cfg.Port,
		logger: logger,
	}
}

func (s *Server) Start() error {
	address := ":" + s.port

	s.logger.Info(
		"http server listening",
		"address",
		address,
	)

	if err := s.app.Listen(address); err != nil {
		return fmt.Errorf("serve http: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info(
		"stopping http server",
	)

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("shutdown http: %w", err)
	}

	return nil
}
