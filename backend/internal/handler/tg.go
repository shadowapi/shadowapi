package handler

import (
	"context"
	ht "github.com/ogen-go/ogen/http"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

func (h *Handler) TgDummy(
	ctx context.Context, req *api.TgSessionCreateReq, params api.TgSessionVerifyParams,
) error {
	//log := h.log.With("handler", "ConnectionSetOAuth2Client")
	//connectionUUID, err := uuid.FromString(params.UUID)
	//if err != nil {
	//	log.Error("failed to parse connection uuid", "error", err)
	//	return ErrWithCode(http.StatusBadRequest, E("failed to parse connection uuid"))
	//}
	//
	//err = query.New(h.dbp).LinkConnectionWithClient(ctx,
	//	query.LinkConnectionWithClientParams{
	//		UUID:     connectionUUID,
	//		ClientID: pgtype.Text{String: req.ClientID, Valid: true},
	//	})
	//if err != nil {
	//	log.Error("failed to link connection with client", "error", err)
	//	return ErrWithCode(http.StatusInternalServerError, E("failed to link connection with client"))
	//}
	return nil
}

// !!!!
// use h.tg. to do telegram things

// TgSessionCreate implements tg-session-create operation.
//
// Create a new Telegram session.
//
// POST /tg
func (h *Handler) TgSessionCreate(ctx context.Context, req *api.TgSessionCreateReq) (*api.Telegram, error) {
	return nil, ht.ErrNotImplemented
}

// TgSessionList implements tg-session-list operation.
//
// List all Telegram sessions for the authenticated user.
//
// GET /tg
func (h *Handler) TgSessionList(ctx context.Context) (*api.TgSessionListOK, error) {
	return nil, ht.ErrNotImplemented
}

// TgSessionVerify implements tg-session-verify operation.
//
// Complete the session creation process by verifying the code.
//
// PUT /tg
func (h *Handler) TgSessionVerify(ctx context.Context, req *api.TgSessionVerifyReq, params api.TgSessionVerifyParams) (*api.Telegram, error) {

	return nil, ht.ErrNotImplemented
}
