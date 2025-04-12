package cmd

import (
	"context"
	"github.com/samber/do/v2"
	"github.com/shadowapi/shadowapi/backend/internal/loader"
	"github.com/spf13/cobra"
	"log/slog"
)

var loaderCmd = &cobra.Command{
	Use:   "loader",
	Short: "loader a yaml and populate datasource, storage, user, etc. for fast testing",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := do.MustInvoke[context.Context](injector)
		loader := do.MustInvoke[*loader.Loader](injector)
		if err := loader.Run(ctx); err != nil {
			slog.Error("failed to start loader", "error", err)
			return
		}
	},
}

func init() {
	LoadDefault(loaderCmd, nil)
	rootCmd.AddCommand(loaderCmd)
}
