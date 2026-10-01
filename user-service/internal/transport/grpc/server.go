package grpc

import (
	"fmt"
	"log/slog"
	"net"

	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"

	googlegrpc "google.golang.org/grpc"
)

type Server struct {
	server *googlegrpc.Server
	port   string
	logger *slog.Logger
}

func NewServer(port string, userService *service.UserService, logger *slog.Logger) *Server {
	grpcServer := googlegrpc.NewServer()

	UserHandler := NewUserHandler(userService)

	userv1.RegisterUserServiceServer(
		grpcServer,
		UserHandler,
	)

	return &Server{
		port:   port,
		server: grpcServer,
		logger: logger,
	}
}

func (s *Server) Start() error {
	address := ":" + s.port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("Listening on port %s, err %w", address, err)
	}

	s.logger.Info("grpc server listening on", "address", address)

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("Error serving: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	s.logger.Info("Server stopping gracefully")

	s.server.GracefulStop()
}
