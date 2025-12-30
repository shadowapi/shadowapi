package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/bridge"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/config"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/kv"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/manager"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/natsconn"
	"github.com/shadowapi/shadowapi/backend/internal/grpc2nats/server"
	"github.com/shadowapi/shadowapi/backend/internal/log"
)

var (
	injector do.Injector
)

var rootCmd = &cobra.Command{
	Use:   "grpc2nats",
	Short: "gRPC to NATS bridge for distributed workers",
	Run:   runServer,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	flags := rootCmd.Flags()
	flags.String("log-level", "info", "log level: debug, info, warn, error")
	if err := viper.BindPFlag("log.level", flags.Lookup("log-level")); err != nil {
		panic(err)
	}
}

func runServer(cmd *cobra.Command, _ []string) {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize dependency injection
	injector = do.New()
	do.ProvideValue(injector, ctx)
	do.Provide(injector, config.Provide)
	do.Provide(injector, log.Provide)
	do.Provide(injector, db.Provide)
	do.Provide(injector, natsconn.Provide)
	do.Provide(injector, kv.Provide)
	do.Provide(injector, manager.Provide)
	do.Provide(injector, server.Provide)
	do.Provide(injector, bridge.Provide)

	logger := do.MustInvoke[*slog.Logger](injector)
	cfg := do.MustInvoke[*config.Config](injector)

	logger.Info("starting grpc2nats service",
		"instance_id", cfg.InstanceID,
		"grpc_port", cfg.GRPC.Port,
	)

	// Start the bridge (subscribes to NATS, manages job flow)
	brg := do.MustInvoke[*bridge.Bridge](injector)
	if err := brg.Start(ctx); err != nil {
		logger.Error("failed to start bridge", "error", err)
		os.Exit(1)
	}

	// Start gRPC server in a goroutine
	grpcSrv := do.MustInvoke[*server.Server](injector)
	go func() {
		if err := grpcSrv.Run(ctx); err != nil {
			logger.Error("gRPC server error", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		logger.Info("received shutdown signal", "signal", sig)
	case <-ctx.Done():
		logger.Info("context cancelled")
	}

	// Graceful shutdown
	logger.Info("shutting down...")
	brg.Stop()
	grpcSrv.Shutdown()

	// Close database pool
	dbp := do.MustInvoke[*pgxpool.Pool](injector)
	if dbp != nil {
		dbp.Close()
	}

	logger.Info("shutdown complete")
}
