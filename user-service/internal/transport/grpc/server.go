package grpc

import (
	"fmt"
	"log/slog"
	"net"

	userv1 "github.com/Ishan-Gijavanekar/user-service/api/proto"
	authv1 "github.com/Ishan-Gijavanekar/user-service/api/proto/auth/v1"
	"github.com/Ishan-Gijavanekar/user-service/internal/auth"
	"github.com/Ishan-Gijavanekar/user-service/internal/service"
	"github.com/Ishan-Gijavanekar/user-service/internal/transport/grpc/interceptor"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server       *googlegrpc.Server
	healthServer *health.Server
	port         string
	logger       *slog.Logger
}

func NewServer(
	port string,
	reflectionEnabled bool,
	userService *service.UserService,
	authService *service.AuthService,
	jwtManager *auth.JWTManager,
	logger *slog.Logger,
) *Server {

	authInterceptor := interceptor.NewAuthInterceptor(
		jwtManager,
		interceptor.PublicMethods,
	)

	grpcServer := googlegrpc.NewServer(
		googlegrpc.ChainUnaryInterceptor(
			interceptor.RequestIDUnaryInterceptor,
			interceptor.LoggingUnaryInterceptor(*logger),
			interceptor.RecoveryUnaryInterceptor(logger),
			authInterceptor.Unary(),
		),
	)

	userHandler := NewUserHandler(
		userService,
	)
	authHandler := NewAuthHandler(
		authService,
	)

	userv1.RegisterUserServiceServer(
		grpcServer,
		userHandler,
	)
	authv1.RegisterAuthServiceServer(
		grpcServer,
		authHandler,
	)

	healthServer := health.NewServer()

	grpc_health_v1.RegisterHealthServer(
		grpcServer,
		healthServer,
	)

	healthServer.SetServingStatus(
		"",
		grpc_health_v1.HealthCheckResponse_SERVING,
	)

	healthServer.SetServingStatus(
		"user.v1.UserService",
		grpc_health_v1.HealthCheckResponse_SERVING,
	)
	healthServer.SetServingStatus(
		"auth.v1.AuthService",
		grpc_health_v1.HealthCheckResponse_SERVING,
	)

	if reflectionEnabled {
		reflection.Register(
			grpcServer,
		)

		logger.Info(
			"grpc reflection enabled",
		)
	}

	return &Server{
		server:       grpcServer,
		healthServer: healthServer,
		port:         port,
		logger:       logger,
	}
}

func (s *Server) Start() error {

	address := ":" + s.port

	listener, err := net.Listen(
		"tcp",
		address,
	)

	if err != nil {
		return fmt.Errorf(
			"listen on %s: %w",
			address,
			err,
		)
	}

	s.logger.Info(
		"grpc server listening",
		"address",
		address,
	)

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf(
			"serve grpc: %w",
			err,
		)
	}

	return nil
}

func (s *Server) Stop() {

	s.logger.Info(
		"stopping grpc server",
	)

	s.healthServer.SetServingStatus(
		"",
		grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	)

	s.healthServer.SetServingStatus(
		"user.v1.UserService",
		grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	)
	s.healthServer.SetServingStatus(
		"auth.v1.AuthService",
		grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	)

	s.server.GracefulStop()
}
