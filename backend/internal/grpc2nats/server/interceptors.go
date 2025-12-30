package server

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor returns a unary server interceptor that logs requests
func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Get client address
		var clientAddr string
		if p, ok := peer.FromContext(ctx); ok {
			clientAddr = p.Addr.String()
		}

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		st, _ := status.FromError(err)

		log.Info("grpc unary request",
			"method", info.FullMethod,
			"client", clientAddr,
			"duration_ms", duration.Milliseconds(),
			"code", st.Code().String(),
		)

		return resp, err
	}
}

// StreamLoggingInterceptor returns a stream server interceptor that logs requests
func StreamLoggingInterceptor(log *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Get client address
		var clientAddr string
		if p, ok := peer.FromContext(ss.Context()); ok {
			clientAddr = p.Addr.String()
		}

		log.Info("grpc stream started",
			"method", info.FullMethod,
			"client", clientAddr,
		)

		// Call the handler
		err := handler(srv, ss)

		// Log the completion
		duration := time.Since(start)
		st, _ := status.FromError(err)

		log.Info("grpc stream ended",
			"method", info.FullMethod,
			"client", clientAddr,
			"duration_s", duration.Seconds(),
			"code", st.Code().String(),
		)

		return err
	}
}
