package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shadowapi/shadowapi/backend/cmd/worker/internal/workerconfig"
)

var (
	injector do.Injector
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "ShadowAPI distributed worker",
	Long:  `A distributed worker that connects to the ShadowAPI backend via gRPC to process jobs.`,
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// LoadWorkerConfig sets up the dependency injector for worker commands
func LoadWorkerConfig(cmd *cobra.Command) {
	cmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		injector = do.New()

		do.ProvideValue(injector, cmd.Context())
		do.Provide(injector, workerconfig.Provide)
		do.Provide(injector, workerconfig.ProvideLogger)
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("WORKER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	pFlags := rootCmd.PersistentFlags()
	pFlags.String("server", "localhost:9090", "gRPC server address (host:port)")
	pFlags.String("log-level", "info", "log level (debug, info, warn, error)")

	if err := viper.BindPFlag("server", pFlags.Lookup("server")); err != nil {
		slog.Error("failed to bind server flag", "error", err)
	}
	if err := viper.BindPFlag("log.level", pFlags.Lookup("log-level")); err != nil {
		slog.Error("failed to bind log-level flag", "error", err)
	}
}
