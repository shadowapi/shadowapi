package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/shadowapi/shadowapi/backend/cmd/worker/internal/workerconfig"
	workerv1 "github.com/shadowapi/shadowapi/backend/pkg/proto/worker/v1"
)

var enrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "Enroll this worker with the backend using a one-time token",
	Long: `Enroll exchanges a one-time enrollment token for permanent worker credentials.
The token is obtained from an administrator who creates it in the backend.

After successful enrollment, save the returned Worker ID and Secret for use
with the 'connect' command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := do.MustInvoke[*workerconfig.Config](injector)
		log := do.MustInvoke[*slog.Logger](injector)

		token, _ := cmd.Flags().GetString("token")
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")

		if token == "" {
			return fmt.Errorf("--token is required")
		}
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		log.Info("connecting to backend for enrollment",
			"server", cfg.Server,
			"name", name,
			"tls", cfg.TLS,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Configure transport credentials based on TLS setting
		var creds credentials.TransportCredentials
		if cfg.TLS {
			creds = credentials.NewTLS(&tls.Config{})
		} else {
			creds = insecure.NewCredentials()
		}

		conn, err := grpc.NewClient(cfg.Server, grpc.WithTransportCredentials(creds))
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer conn.Close()

		client := workerv1.NewWorkerServiceClient(conn)

		resp, err := client.Enroll(ctx, &workerv1.EnrollRequest{
			EnrollmentToken: token,
			WorkerName:      name,
			WorkerVersion:   version,
		})
		if err != nil {
			return fmt.Errorf("enrollment failed: %w", err)
		}

		fmt.Println()
		fmt.Println("=== Enrollment Successful ===")
		fmt.Println()
		fmt.Printf("Worker ID:     %s\n", resp.WorkerId)
		fmt.Printf("Worker Secret: %s\n", resp.WorkerSecret)
		fmt.Printf("Is Global:     %v\n", resp.IsGlobal)
		fmt.Printf("Workspaces:    %v\n", resp.AllowedWorkspaces)
		fmt.Println()
		fmt.Println("=== IMPORTANT ===")
		fmt.Println("Save these credentials! The secret cannot be retrieved again.")
		fmt.Println()
		fmt.Println("To connect, set these environment variables:")
		fmt.Printf("  export WORKER_ID=%s\n", resp.WorkerId)
		fmt.Printf("  export WORKER_SECRET=%s\n", resp.WorkerSecret)
		fmt.Printf("  export WORKER_SERVER=%s\n", cfg.Server)
		if cfg.TLS {
			fmt.Println("  export WORKER_TLS=true")
		}
		fmt.Println()
		fmt.Println("Then run: worker connect")

		return nil
	},
}

func init() {
	LoadWorkerConfig(enrollCmd)
	enrollCmd.Flags().String("token", "", "One-time enrollment token from administrator (required)")
	enrollCmd.Flags().String("name", "", "Display name for this worker (required)")
	enrollCmd.Flags().String("version", "1.0.0", "Worker version string")
	rootCmd.AddCommand(enrollCmd)
}
