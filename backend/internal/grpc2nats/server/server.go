package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"google.golang.org/grpc"

	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/manager"
	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

// Server is the gRPC server for distributed workers
type Server struct {
	cfg      *config.Config
	log      *slog.Logger
	dbp      *pgxpool.Pool
	grpcSrv  *grpc.Server
	listener net.Listener
	manager  *manager.WorkerManager
}

// Provide creates a new gRPC server for the dependency injector
func Provide(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*slog.Logger](i).With("component", "grpc-server")
	dbp := do.MustInvoke[*pgxpool.Pool](i)
	mgr := do.MustInvoke[*manager.WorkerManager](i)

	// Create gRPC server with interceptors
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor(log),
		),
		grpc.ChainStreamInterceptor(
			StreamLoggingInterceptor(log),
		),
	)

	// Register the worker service
	workerSvc := NewWorkerService(log, dbp, mgr)
	workerv1.RegisterWorkerServiceServer(grpcSrv, workerSvc)

	log.Info("gRPC server initialized",
		"host", cfg.GRPC.Host,
		"port", cfg.GRPC.Port,
	)

	return &Server{
		cfg:     cfg,
		log:     log,
		dbp:     dbp,
		grpcSrv: grpcSrv,
		manager: mgr,
	}, nil
}

// Run starts the gRPC server
func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.GRPC.Host, s.cfg.GRPC.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.Error("failed to listen", "address", addr, "error", err)
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.log.Info("gRPC server starting", "address", addr)
	return s.grpcSrv.Serve(listener)
}

// Shutdown gracefully stops the gRPC server
func (s *Server) Shutdown() {
	s.log.Info("gRPC server shutting down")
	s.grpcSrv.GracefulStop()
	if s.listener != nil {
		s.listener.Close()
	}
}

// Manager returns the worker manager for external access
func (s *Server) Manager() *manager.WorkerManager {
	return s.manager
}
