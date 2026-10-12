package http

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	httpmiddleware "github.com/Ishan-Gijavanekar/user-service/internal/transport/http/middleware"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *slog.Logger
}

func NewServer(
	port string,
	userService *service.UserService,
	authService *service.AuthService,
	jwtManager *auth.JWTManager,
	logger *slog.Logger,
) *Server {
	app := fiber.New()

	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(userService)
	authMiddleware := httpmiddleware.NewAuthMiddleware(jwtManager)

	RegisterRoutes(
		app,
		authHandler,
		userHandler,
		authMiddleware,
	)

	return &Server{
		app:    app,
		port:   port,
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
