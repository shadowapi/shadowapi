package jobs

/*
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
	"log/slog"
)

type emailOAuthSettings struct {
	OAuth2ClientUUID string `json:"oauth2_client_uuid"`
	OAuth2TokenUUID  string `json:"oauth2_token_uuid"`
}

func fetchGmailEmails(ctx context.Context, ds query.Datasource, since time.Time, dbp *pgxpool.Pool, log *slog.Logger) ([]api.Message, error) {
	// Parse the datasource settings for OAuth2 credentials.
	if len(ds.Settings) == 0 {
		return nil, errors.New("empty datasource settings")
	}
	var settings emailOAuthSettings
	if err := json.Unmarshal(ds.Settings, &settings); err != nil {
		log.Error("failed to unmarshal datasource settings", "error", err)
		return nil, err
	}
	if settings.OAuth2ClientUUID == "" || settings.OAuth2TokenUUID == "" {
		return nil, errors.New("missing OAuth2 configuration in datasource settings")
	}
	// Get the OAuth2 client configuration.
	clientConfig, err := oauth2.GetClientConfig(ctx, dbp, settings.OAuth2ClientUUID)
	if err != nil {
		log.Error("failed to get client config", "error", err)
		return nil, err
	}
	// Convert token UUID string to uuid.UUID.
	tokenUUID, err := uuid.FromString(settings.OAuth2TokenUUID)
	if err != nil {
		log.Error("invalid OAuth2 token UUID", "token_uuid", settings.OAuth2TokenUUID, "error", err)
		return nil, err
	}
	// Create a token store and get a valid token.
	store, err := newTokenStore(ctx, clientConfig, dbp, tokenUUID, time.Minute)
	if err != nil {
		log.Error("failed to create token store", "error", err)
		return nil, err
	}
	token, err := store.Token()
	if err != nil {
		log.Error("failed to retrieve OAuth2 token", "error", err)
		return nil, err
	}
	// Create an HTTP client using the token.
	httpClient := clientConfig.Client(ctx, token)
	// Create the Gmail API service.
	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		log.Error("failed to create Gmail service", "error", err)
		return nil, err
	}
	// Build a query string for messages after the given date.
	queryStr := fmt.Sprintf("after:%s", since.Format("2006/01/02"))
	req := gmailService.Users.Messages.List("me").Q(queryStr)
	res, err := req.Do()
	if err != nil {
		log.Error("failed to list messages", "error", err, "query", queryStr)
		return nil, err
	}
	if res.ResultSizeEstimate == 0 || len(res.Messages) == 0 {
		return []api.Message{}, nil
	}
	var messages []api.Message
	// Helper function to extract header values.
	getHeader := func(headers []*gmail.MessagePartHeader, name string) string {
		for _, h := range headers {
			if strings.EqualFold(h.Name, name) {
				return h.Value
			}
		}
		return ""
	}
	// Loop over each message and fetch its full content.
	for _, m := range res.Messages {
		fullMsg, err := gmailService.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			log.Error("failed to get full message", "message_id", m.Id, "error", err)
			continue
		}
		var subjectOpt api.OptString
		subject := getHeader(fullMsg.Payload.Headers, "Subject")
		if subject != "" {
			subjectOpt = api.NewOptString(subject)
		} else {
			subjectOpt = api.NewOptString("")
		}
		// Prepare the api.Message. Using snippet as the body.
		emailMsg := api.Message{
			UUID:       api.NewOptString("gmail-" + m.Id),
			Format:     "email",
			Type:       "email",
			Sender:     getHeader(fullMsg.Payload.Headers, "From"),
			Recipients: []string{getHeader(fullMsg.Payload.Headers, "To")},
			Subject:    subjectOpt,
			Body:       fullMsg.Snippet,
		}
		messages = append(messages, emailMsg)
	}
	log.Info("Gmail fetch completed", "count", len(messages))
	return messages, nil
}
*/
