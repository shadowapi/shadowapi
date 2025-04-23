package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	"github.com/shadowapi/shadowapi/backend/internal/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts UI and RESTfull API servers",
	Run: func(cmd *cobra.Command, args []string) {
		// DI must know all modules
		// injector - DI god-like object, instance of all modules
		// injector is ...
		ctx := do.MustInvoke[context.Context](injector)
		srv := do.MustInvoke[*server.Server](injector)
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
	},
}

func init() {
	LoadDefault(serveCmd, nil)
	rootCmd.AddCommand(serveCmd)
}
