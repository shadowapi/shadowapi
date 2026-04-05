package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	oauth2tools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Pipeline operations",
}

// ── list ─────────────────────────────────────────────────────────────

var pipelineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pipelines",
	Run: func(cmd *cobra.Command, args []string) {
		dbp := do.MustInvoke[*pgxpool.Pool](injector)
		q := query.New(dbp)

		rows, err := q.GetPipelines(cmd.Context(), query.GetPipelinesParams{
			OrderBy: "created_at", OrderDirection: "desc",
			Limit: 50, Type: "", IsEnabled: -1,
		})
		if err != nil {
			slog.Error("failed to list pipelines", "error", err)
			return
		}
		if len(rows) == 0 {
			fmt.Println("No pipelines found.")
			return
		}
		fmt.Printf("%-38s %-25s %-12s %-8s\n", "UUID", "Name", "Type", "Enabled")
		fmt.Println(strings.Repeat("─", 90))
		for _, r := range rows {
			fmt.Printf("%-38s %-25s %-12s %-8v\n",
				r.UUID.String(), r.Name, r.Type, r.IsEnabled)
		}
	},
}

// ── run ──────────────────────────────────────────────────────────────

var pipelineRunLimit int64

var pipelineRunCmd = &cobra.Command{
	Use:   "run <pipeline-uuid>",
	Short: "Fetch messages for a pipeline and store them (direct, no NATS)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		dbp := do.MustInvoke[*pgxpool.Pool](injector)
		log := do.MustInvoke[*slog.Logger](injector)

		pipeUUID, err := uuid.FromString(args[0])
		if err != nil {
			slog.Error("invalid pipeline UUID", "error", err)
			return
		}

		msgs, err := runPipelineFetch(ctx, log, dbp, pipeUUID, pipelineRunLimit)
		if err != nil {
			slog.Error("pipeline run failed", "error", err)
			return
		}
		fmt.Printf("\nFetched and stored %d messages.\n", len(msgs))
	},
}

func runPipelineFetch(ctx context.Context, log *slog.Logger, dbp *pgxpool.Pool, pipeUUID uuid.UUID, limit int64) ([]api.Message, error) {
	q := query.New(dbp)

	// 1. Load pipeline
	pipeRow, err := q.GetPipeline(ctx, pgtype.UUID{Bytes: pipeUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}
	fmt.Printf("Pipeline: %s (%s, type=%s, enabled=%v)\n",
		pipeRow.Pipeline.Name, pipeRow.Pipeline.UUID, pipeRow.Pipeline.Type, pipeRow.Pipeline.IsEnabled)

	// 2. Load datasource
	dsRow, err := q.GetDatasource(ctx, converter.UuidPtrToPgUUID(pipeRow.Pipeline.DatasourceUUID))
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}
	fmt.Printf("Datasource: %s (type=%s, enabled=%v)\n",
		dsRow.Datasource.UUID, dsRow.Datasource.Type, dsRow.Datasource.IsEnabled)

	if dsRow.Datasource.Type != "email_oauth" {
		return nil, fmt.Errorf("unsupported datasource type: %s (only email_oauth supported)", dsRow.Datasource.Type)
	}

	// 3. Parse datasource settings
	var cfg api.DatasourceEmailOAuth
	if err := json.Unmarshal(dsRow.Datasource.Settings, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse datasource settings: %w", err)
	}
	if cfg.OAuth2ClientUUID == "" {
		return nil, errors.New("datasource settings missing oauth2_client_uuid")
	}
	fmt.Printf("OAuth2 Client: %s\n", cfg.OAuth2ClientUUID)

	// 4. Get OAuth2 client config
	clientConfig, err := oauth2tools.GetClientConfig(ctx, dbp, cfg.OAuth2ClientUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 client config: %w", err)
	}
	fmt.Printf("OAuth2 Provider: %s\n", clientConfig.Provider)

	// 5. Get token
	pgClientUUID, _ := converter.ConvertStringToPgUUID(cfg.OAuth2ClientUUID)
	tokenRows, err := q.GetOauth2TokensByClientUUID(ctx, pgClientUUID)
	if err != nil || len(tokenRows) == 0 {
		return nil, fmt.Errorf("no OAuth2 token found for client %s", cfg.OAuth2ClientUUID)
	}
	latest := tokenRows[0]
	tokenUUID := latest.Oauth2Token.UUID
	fmt.Printf("Token UUID: %s\n", tokenUUID)

	// 6. Build token store (auto-refreshes and persists)
	tokenStore, err := oauth2tools.NewTokenStore(ctx, &clientConfig.Config, dbp, tokenUUID, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to create token store: %w", err)
	}
	token, err := tokenStore.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	fmt.Printf("Token valid until: %s\n", token.Expiry.Format(time.RFC3339))

	// 7. Build Gmail service
	httpClient := clientConfig.Config.Client(ctx, token)
	gmailSvc, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// 8. Fetch message IDs
	fmt.Printf("\nFetching up to %d messages from Gmail...\n", limit)
	listRes, err := gmailSvc.Users.Messages.List("me").MaxResults(limit).Do()
	if err != nil {
		return nil, fmt.Errorf("Gmail API list failed: %w", err)
	}
	if len(listRes.Messages) == 0 {
		fmt.Println("No messages found.")
		return nil, nil
	}
	fmt.Printf("Found %d message IDs.\n\n", len(listRes.Messages))

	// 9. Fetch full messages and store
	headerVal := func(headers []*gmail.MessagePartHeader, name string) string {
		for _, h := range headers {
			if strings.EqualFold(h.Name, name) {
				return h.Value
			}
		}
		return ""
	}

	var messages []api.Message
	for i, meta := range listRes.Messages {
		fullMsg, err := gmailSvc.Users.Messages.Get("me", meta.Id).Format("full").Do()
		if err != nil {
			log.Warn("failed to fetch message", "id", meta.Id, "error", err)
			continue
		}

		subject := headerVal(fullMsg.Payload.Headers, "Subject")
		from := headerVal(fullMsg.Payload.Headers, "From")
		to := headerVal(fullMsg.Payload.Headers, "To")

		// Generate deterministic UUID from gmail message ID
		msgUUID := uuid.NewV5(uuid.NamespaceURL, "gmail-"+meta.Id)

		// Store in message table
		_, err = q.CreateMessage(ctx, query.CreateMessageParams{
			UUID:       pgtype.UUID{Bytes: msgUUID, Valid: true},
			Sender:     from,
			Recipients: []string{to},
			Subject:    converter.PgText(subject),
			Body:       fullMsg.Snippet,
			Format:     "email",
			Type:       "email",
		})
		if err != nil {
			// likely duplicate — skip
			log.Debug("insert skipped (duplicate?)", "id", meta.Id, "error", err)
		}

		msg := api.Message{
			UUID:       api.NewOptString(msgUUID.String()),
			Type:       "email",
			Format:     "email",
			Sender:     from,
			Recipients: []string{to},
			Subject:    api.NewOptString(subject),
			Body:       fullMsg.Snippet,
		}
		messages = append(messages, msg)

		fmt.Printf("  %2d. %s — %s\n", i+1, from, subject)
	}

	return messages, nil
}

func init() {
	pipelineRunCmd.Flags().Int64VarP(&pipelineRunLimit, "limit", "l", 10, "max messages to fetch")
	pipelineCmd.AddCommand(pipelineListCmd)
	pipelineCmd.AddCommand(pipelineRunCmd)

	LoadDefault(pipelineCmd, nil)
	rootCmd.AddCommand(pipelineCmd)
}
