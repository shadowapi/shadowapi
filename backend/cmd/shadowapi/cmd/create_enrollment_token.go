package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

var createEnrollmentTokenCmd = &cobra.Command{
	Use:   "create-enrollment-token",
	Short: "Create a worker enrollment token for bootstrap",
	Long: `Creates a one-time enrollment token for a distributed worker.
The token is output to stdout for script capture.

This command uses BE_INIT_ADMIN_EMAIL to look up the admin user UUID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		isGlobal, _ := cmd.Flags().GetBool("global")
		expiresInHours, _ := cmd.Flags().GetInt("expires-in")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		ctx := do.MustInvoke[context.Context](injector)
		dbp := do.MustInvoke[*pgxpool.Pool](injector)
		log := do.MustInvoke[*slog.Logger](injector)

		// Look up admin user by email from environment
		adminEmail := os.Getenv("BE_INIT_ADMIN_EMAIL")
		if adminEmail == "" {
			return fmt.Errorf("BE_INIT_ADMIN_EMAIL environment variable not set")
		}

		q := query.New(dbp)

		// Get admin user UUID
		user, err := q.GetUserByEmail(ctx, adminEmail)
		if err != nil {
			return fmt.Errorf("admin user not found: %w", err)
		}

		// Generate secure random token (32 bytes = 64 hex chars)
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return fmt.Errorf("failed to generate token: %w", err)
		}
		rawToken := hex.EncodeToString(tokenBytes)

		// Hash the token for storage using SHA256 (must match gRPC service validation)
		hasher := sha256.New()
		hasher.Write([]byte(rawToken))
		tokenHash := hex.EncodeToString(hasher.Sum(nil))

		// Set expiration
		expiresAt := time.Now().Add(time.Duration(expiresInHours) * time.Hour)

		tokenUUID := uuid.Must(uuid.NewV7())

		_, err = q.CreateEnrollmentToken(ctx, query.CreateEnrollmentTokenParams{
			UUID:              tokenUUID,
			TokenHash:         tokenHash,
			Name:              name,
			IsGlobal:          isGlobal,
			WorkspaceUuids:    []pgtype.UUID{}, // Empty for global workers
			ExpiresAt:         pgtype.Timestamptz{Time: expiresAt, Valid: true},
			CreatedByUserUuid: &user.UUID,
		})
		if err != nil {
			return fmt.Errorf("failed to create token: %w", err)
		}

		log.Info("enrollment token created",
			"uuid", tokenUUID,
			"name", name,
			"is_global", isGlobal,
			"expires_at", expiresAt,
		)

		// Output only the raw token to stdout for script capture
		fmt.Println(rawToken)
		return nil
	},
}

func init() {
	LoadDefault(createEnrollmentTokenCmd, nil)
	createEnrollmentTokenCmd.Flags().String("name", "", "Name for the enrollment token (required)")
	createEnrollmentTokenCmd.Flags().Bool("global", true, "Whether worker can access all workspaces")
	createEnrollmentTokenCmd.Flags().Int("expires-in", 1, "Token expiration in hours")
	rootCmd.AddCommand(createEnrollmentTokenCmd)
}
