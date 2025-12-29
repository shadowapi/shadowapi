package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	grpcserver "github.com/shadowapi/shadowapi/backend/internal/grpc"
	"github.com/shadowapi/shadowapi/backend/internal/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts UI, RESTfull API, and gRPC servers",
	Run: func(cmd *cobra.Command, args []string) {
		// DI must know all modules
		// injector - DI god-like object, instance of all modules
		ctx := do.MustInvoke[context.Context](injector)
		srv := do.MustInvoke[*server.Server](injector)
		grpcSrv := do.MustInvoke[*grpcserver.Server](injector)

		// Start gRPC server in a goroutine
		go func() {
			if err := grpcSrv.Run(ctx); err != nil {
				slog.Error("failed to start gRPC server", "error", err)
			}
		}()

		// Start HTTP server (blocking)
		if err := srv.Run(ctx); err != nil {
			slog.Error("failed to start server", "error", err)
			return
		}

		// wait for shutdown signal
		s, err := injector.RootScope().ShutdownOnSignals(os.Interrupt)
		if err != nil {
			slog.Error("failed to shutdown on signals", "error", err.Error(), "signal", s)
			return
		}

		// Graceful shutdown
		grpcSrv.Shutdown()
	},
}

func init() {
	LoadDefault(serveCmd, nil)
	rootCmd.AddCommand(serveCmd)
}
