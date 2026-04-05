package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run health checks (db, storages, tokens, pipelines)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		dbp := do.MustInvoke[*pgxpool.Pool](injector)
		q := query.New(dbp)

		fmt.Println("ShadowAPI Health Check")
		fmt.Println(strings.Repeat("═", 50))

		// DB
		var one int
		err := dbp.QueryRow(ctx, "SELECT 1").Scan(&one)
		printCheck("PostgreSQL", err)

		// Storages
		fmt.Println("\n── Storages ──")
		storages, err := q.GetStorages(ctx, query.GetStoragesParams{
			OrderBy: "created_at", OrderDirection: "desc",
			Limit: 50, IsEnabled: -1,
		})
		if err != nil {
			fmt.Printf("  FAIL: %v\n", err)
		} else if len(storages) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, s := range storages {
				status := "disabled"
				if s.IsEnabled {
					status = "enabled"
				}
				fmt.Printf("  %-30s %-10s %s\n", s.Name, s.Type, status)
			}
		}

		// Datasources
		fmt.Println("\n── Datasources ──")
		datasources, err := q.GetDatasources(ctx, query.GetDatasourcesParams{
			OrderBy: "created_at", OrderDirection: "desc",
			Limit: 50, IsEnabled: -1,
		})
		if err != nil {
			fmt.Printf("  FAIL: %v\n", err)
		} else if len(datasources) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, d := range datasources {
				status := "disabled"
				if d.IsEnabled {
					status = "enabled"
				}
				fmt.Printf("  %-38s %-15s %-12s %s\n",
					d.UUID, d.Name, d.Type, status)
			}
		}

		// Pipelines
		fmt.Println("\n── Pipelines ──")
		pipes, err := q.GetPipelines(ctx, query.GetPipelinesParams{
			OrderBy: "created_at", OrderDirection: "desc",
			Limit: 50, Type: "", IsEnabled: -1,
		})
		if err != nil {
			fmt.Printf("  FAIL: %v\n", err)
		} else if len(pipes) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, p := range pipes {
				status := "disabled"
				if p.IsEnabled {
					status = "enabled"
				}
				fmt.Printf("  %-38s %-25s %-12s %s\n",
					p.UUID, p.Name, p.Type, status)
			}
		}

		// Schedulers
		fmt.Println("\n── Schedulers ──")
		schedulers, err := q.GetSchedulers(ctx, query.GetSchedulersParams{
			OrderBy: "created_at", OrderDirection: "desc",
			Limit: 50, IsEnabled: -1, IsPaused: -1,
		})
		if err != nil {
			fmt.Printf("  FAIL: %v\n", err)
		} else if len(schedulers) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, s := range schedulers {
				nextRun := "never"
				if s.NextRun.Valid {
					nextRun = s.NextRun.Time.Format(time.RFC3339)
				}
				lastRun := "never"
				if s.LastRun.Valid {
					lastRun = s.LastRun.Time.Format(time.RFC3339)
				}
				fmt.Printf("  %-38s cron=%-15s next=%s last=%s enabled=%v\n",
					s.UUID, s.CronExpression.String, nextRun, lastRun, s.IsEnabled)
			}
		}

		// OAuth2 Tokens
		fmt.Println("\n── OAuth2 Tokens ──")
		tokens, err := q.GetOauth2Tokens(ctx, query.GetOauth2TokensParams{
			OrderBy: "created_at", OrderDirection: "desc",
			Limit: 50, ClientUuid: "",
		})
		if err != nil {
			fmt.Printf("  FAIL: %v\n", err)
		} else if len(tokens) == 0 {
			fmt.Println("  (none)")
		} else {
			now := time.Now()
			for _, t := range tokens {
				status := "unknown"
				if t.UpdatedAt.Valid {
					age := now.Sub(t.UpdatedAt.Time)
					status = fmt.Sprintf("updated %s ago", age.Truncate(time.Minute))
				}
				clientStr := ""
				if t.ClientUuid != nil {
					clientStr = t.ClientUuid.String()
				}
				fmt.Printf("  %-38s client=%-38s %s\n",
					t.UUID, clientStr, status)
			}
		}

		// Messages count
		fmt.Println("\n── Messages ──")
		var msgCount int64
		err = dbp.QueryRow(ctx, "SELECT count(*) FROM message").Scan(&msgCount)
		if err != nil {
			fmt.Printf("  FAIL: %v\n", err)
		} else {
			fmt.Printf("  Total: %d\n", msgCount)
		}

		fmt.Println()
	},
}

func printCheck(name string, err error) {
	if err != nil {
		fmt.Printf("  FAIL  %s: %v\n", name, err)
	} else {
		fmt.Printf("  OK    %s\n", name)
	}
}

func init() {
	LoadDefault(checkCmd, nil)
	rootCmd.AddCommand(checkCmd)
}
