package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shadowapi/shadowapi/backend/internal/auth"
	"github.com/shadowapi/shadowapi/backend/internal/config"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/handler"
	"github.com/shadowapi/shadowapi/backend/internal/loader"
	"github.com/shadowapi/shadowapi/backend/internal/log"
	"github.com/shadowapi/shadowapi/backend/internal/queue"
	"github.com/shadowapi/shadowapi/backend/internal/server"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/internal/worker"
)

var (
	// defaultConfigPath is the default path to the config file
	defaultConfigPath string

	// injector is the dependency injector
	injector do.Injector
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "shadowapi",
	Short: "synchronize, transform and search your emails,messages and social media",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// LoadDefault loads default config and database connection.
func LoadDefault(cmd *cobra.Command, modify func(cfg *config.Config)) {
	cmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		injector = do.New()
		do.ProvideNamedValue(injector, "defaultConfigPath", defaultConfigPath)

		do.ProvideValue(injector, cmd.Context())
		do.Provide(injector, config.Provide)
		do.Provide(injector, log.Provide)
		do.Provide(injector, db.Provide)
		do.Provide(injector, loader.Provide)

		fmt.Println("cmd.Name:", cmd.Name())
		fmt.Println("cmd.Use:", cmd.Use)
		fmt.Println("SA_SKIP_WORKER:", os.Getenv("SA_SKIP_WORKER"))

		// Skip server when subcommand is loader
		if cmd.Name() != "loader" {
			do.Provide(injector, queue.Provide)
			do.Provide(injector, auth.Provide)
			do.Provide(injector, session.Provide)
			do.Provide(injector, handler.Provide)
			do.Provide(injector, server.Provide)
		}

		if os.Getenv("SA_SKIP_WORKER") == "true" {
			do.Provide(injector, worker.ProvideLazy)
		} else {
			do.Provide(injector, worker.Provide)
		}

		if modify != nil {
			modify(do.MustInvoke[*config.Config](injector))
		}

		////---------------------------------------
		//// Provide dynamic connections
		////---------------------------------------
		//// Build the map[string]*pgxpool.Pool from storages table
		//pgConns, err := storages.ProvideDynamicPGConnections(injector)
		//if err != nil {
		//	// handle error, or log/fatal
		//	panic(err) // or log and exit
		//}
		//// Register so other code can do: do.MustInvoke[map[string]*pgxpool.Pool](injector)
		//do.ProvideValue(injector, pgConns)
		//
		//// Build the map[string]*s3.S3 from storages table
		//s3Conns, err := storages.ProvideDynamicS3Connections(injector)
		//if err != nil {
		//	// handle error
		//	panic(err)
		//}
		//do.ProvideValue(injector, s3Conns)
		//
		////---------------------------------------
		//// Start the Postgres reconnect routine
		////---------------------------------------
		//storages.ReconnectDynamicPGConnections(cmd.Context(), pgConns, injector)

	}
	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		// Close the database connection pool
		// only when the pool has actually been created, as some commands
		// create fake database connections just to satisfy the dependency
		dbPool := do.MustInvoke[*pgxpool.Pool](injector)
		if dbPool != nil {
			dbPool.Close()
		}
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	pFlags := rootCmd.PersistentFlags()
	pFlags.StringVar(&defaultConfigPath, "config", "config.yaml", "config file")

	pFlags.String("log-level", "info", "log level, one of: debug, info, warn, error")
	if err := viper.BindPFlag("log.level", pFlags.Lookup("log-level")); err != nil {
		panic(err.Error())
	}

	flags := rootCmd.Flags()
	flags.BoolP("toggle", "t", false, "Help message for toggle")
}
