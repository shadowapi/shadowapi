package worker

/*
// OLD CODE
import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	oauthTools "github.com/shadowapi/shadowapi/backend/internal/oauth2"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Args for the token refresh worker
type tokenRefresherWorkerArgs struct {
	Expiry    time.Time
	TokenUUID uuid.UUID
}

// Worker which refreshes the token
type tokenRefresherWorker struct {
	dbp *pgxpool.Pool
	log *slog.Logger
}

// Work to refresh the token
//
// 1. Get the token by UUID from the database
// 2. If token doesn't exists - cancel the worker
// 3. Unmarshal the token data
// 4. Start token refresh process
// 5. Schedule the next token refresh according to the token expiry
func (t *tokenRefresherWorker) Work(ctx context.Context, b *Broker, args *tokenRefresherWorkerArgs) error {
	log := t.log.With("token_uuid", args.TokenUUID, "worker", "tokenRefresherWorker")
	log.Debug("start token refresh worker")
	token, err := db.InTx(ctx, t.dbp, func(tx pgx.Tx) (*oauth2.Token, error) {
		db := query.New(t.dbp).WithTx(tx)
		tokenData, err := db.GetOauth2TokenByUUID(ctx, args.TokenUUID)
		if err != nil {
			return nil, err
		}

		var token oauth2.Token
		if err = json.Unmarshal(tokenData.Token, &token); err != nil {
			log.Error("fail unmarshal token data", "error", err)
			return nil, err
		}

		config, err := oauthTools.GetClientConfig(ctx, t.dbp, tokenData.ClientID)
		if err != nil {
			log.Error("failed get OAuth2 client config", "error", err)
			return nil, err
		}

		rawToken, err := config.TokenSource(ctx, &token).Token()
		if err != nil {
			log.Error("failed refresh token, delete it", "error", err)
			if err := db.DeleteOauth2Token(ctx, args.TokenUUID); err != nil {
				log.Error("failed delete broken token", "error", err)
			}
			return nil, err
		}
		return rawToken, nil
	})
	if err == pgx.ErrNoRows {
		log.Warn("token not found, cancel refresh worker")
		return nil
	} else if err != nil {
		return err
	}

	log.Debug("token refreshed successfully, ready to schedule next refresh")
	// reschedule the worker
	return b.ScheduleRefresh(ctx, args.TokenUUID, token.Expiry)
}
*/
