package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"net/http"
)

func (h *Handler) StorageHostfilesCreate(ctx context.Context, req *api.StorageHostfiles) (*api.StorageHostfiles, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) StorageHostfilesDelete(ctx context.Context, params api.StorageHostfilesDeleteParams) error {
	//TODO implement me
	return ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) StorageHostfilesGet(ctx context.Context, params api.StorageHostfilesGetParams) (*api.StorageHostfiles, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) StorageHostfilesUpdate(ctx context.Context, req *api.StorageHostfiles, params api.StorageHostfilesUpdateParams) (*api.StorageHostfiles, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}
