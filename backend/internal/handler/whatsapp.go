package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"net/http"
)

func (h *Handler) WhatsappContacts(ctx context.Context) (*api.WhatsappContactsOK, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) WhatsappDownloadAttachment(ctx context.Context, req *api.WhatsappDownloadAttachmentReq) (*api.WhatsappDownloadAttachmentOK, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) WhatsappDownloadMessage(ctx context.Context, req *api.WhatsappDownloadMessageReq) (*api.WhatsappDownloadMessageOK, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) WhatsappLogin(ctx context.Context) (*api.WhatsAppLoginResponse, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) WhatsappStatus(ctx context.Context) (*api.WhatsAppStatusResponse, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) WhatsappSync(ctx context.Context, req *api.WhatsappSyncReq) (*api.WhatsappSyncOK, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}
