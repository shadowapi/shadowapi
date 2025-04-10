package cmd

import (
	"context"
	"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/server"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var loaderCmd = &cobra.Command{
	Use:   "loader",
	Short: "loader a yaml and populate datasource, storage, user, etc. for fast testing",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := do.MustInvoke[context.Context](injector)
		srv := do.MustInvoke[*server.Server](injector)
		if err := srv.Run(ctx); err != nil {
			slog.Error("failed to start server", "error", err)
			return
		}
		s, err := injector.RootScope().ShutdownOnSignals(os.Interrupt)
		if err != nil {
			slog.Error("failed to shutdown on signals", "error", err.Error(), "signal", s)
			return
		}
	},
}

func init() {
	LoadDefault(loaderCmd, nil)
	rootCmd.AddCommand(loaderCmd)
}
