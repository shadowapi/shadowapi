package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password [email] [new-password]",
	Short: "Reset password for a local (non-ZITADEL) admin user",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]
		pass := args[1]
		ctx := do.MustInvoke[context.Context](injector)
		dbp := do.MustInvoke[*pgxpool.Pool](injector)
		log := do.MustInvoke[*slog.Logger](injector)

		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		res, err := dbp.Exec(ctx, `UPDATE "user" SET password=$1, updated_at=NOW() WHERE email=$2 AND (zitadel_subject IS NULL OR zitadel_subject='')`, string(hash), email)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("user not found or managed by ZITADEL")
		}
		log.Info("password reset", "email", email)
		return nil
	},
}

func init() {
	LoadDefault(resetPasswordCmd, nil)
	rootCmd.AddCommand(resetPasswordCmd)
}
