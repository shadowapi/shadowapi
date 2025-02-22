package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
	"github.com/jackc/pgx/v5"
	"github.com/shadowapi/shadowapi/backend/internal/tg/telegram"

	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// handleListSessions handles the GET /v1/sessions endpoint
func (h *Handler) handleListSessions(ctx context.Context, _ *Empty) (*SessionListResponse, error) {

	var accountID int64 = 1

	tx := query.New(h.dbp)
	list, err := tx.TgGetSessionList(ctx, accountID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return &SessionListResponse{
			Body: SessionListResponseBody{
				Total: 0,
			},
		}, nil
	}

	response := &SessionListResponse{
		SessionListResponseBody{
			Total:    int64(len(list)),
			Sessions: make([]Session, 0, len(list)),
		},
	}

	for _, item := range list {
		session := Session{
			ID:          item.ID,
			Phone:       item.Phone,
			Description: item.Description.String,
			UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		}

		response.Body.Sessions = append(response.Body.Sessions, session)
	}

	return response, nil
}

// handleSignIn handles the POST /v1/sessions endpoint
func (h *Handler) handleCreateSessionStart(ctx context.Context, req *SignInRequest) (*SignInResponse, error) {
	var accountID int64 = 1

	var session int64
	tx := query.New(h.dbp)

	// create session with phone number
	session, err := tx.TgCreateSession(ctx, query.TgCreateSessionParams{
		AccountID: accountID,
		Phone:     req.Body.Phone,
	})

	if err != nil {
		return nil, err
	}

	client := h.client(session)
	sentCode, err := client.SendCode(ctx, req.Body.Phone)
	if err != nil {
		return nil, err
	}

	return &SignInResponse{
		Body: SignInResponseBody{
			SessionID:     session,
			PhoneCodeHash: sentCode.PhoneCodeHash,
			Timeout:       int64(sentCode.Timeout),
			NextType:      fmt.Sprintf("%T", sentCode.Type),
		},
	}, nil
}

// handleSendCode handles the PUT /v1/sessions/{id} endpoint
func (h *Handler) handleCreateSessionFin(ctx context.Context, req *SendCodeRequest) (*SendCodeResponse, error) {
	tx := query.New(h.dbp)

	session, err := tx.TgGetSession(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	client := h.client(req.ID)
	auth, err := client.SignIn(ctx, session.Phone, req.Body.Code, req.Body.Password, req.Body.PhoneCodeHash)
	if err != nil {
		return nil, err
	}

	switch v := auth.User.(type) {
	case *tg.User:
		return &SendCodeResponse{
			Body: User{
				ID:        v.ID,
				Username:  v.Username,
				FirstName: v.FirstName,
				LastName:  v.LastName,
				Phone:     v.Phone,
			},
		}, nil
	default:
		return nil, fmt.Errorf("sign in error: unknown user type %T", auth.User)
	}
}

// handleSelf handles the GET /v1/sessions/{id}/self endpoint
func (h *Handler) handleSelf(ctx context.Context, req *SelfRequest) (*SelfResponse, error) {

	tx := query.New(h.dbp)
	if _, err := tx.TgGetSession(ctx, req.ID); err != nil {
		return nil, err
	}

	client := h.client(req.ID)

	self, err := client.Self(ctx)
	if err != nil {
		return nil, err
	}

	return &SelfResponse{
		Body: User{
			ID:        self.ID,
			Username:  self.Username,
			FirstName: self.FirstName,
			LastName:  self.LastName,
			Phone:     self.Phone,
		},
	}, nil
}

func (h *Handler) client(session int64) *telegram.Telegram {
	cfg := h.cfg.Telegram
	return telegram.New(cfg.AppID, cfg.AppHash, telegram.WithSessionStorage(session, h.dbp))
}
