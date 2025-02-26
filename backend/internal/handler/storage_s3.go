package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"net/http"
)

func (h *Handler) StorageS3Create(ctx context.Context, req *api.StorageS3) (*api.StorageS3, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) StorageS3Delete(ctx context.Context, params api.StorageS3DeleteParams) error {
	//TODO implement me
	return ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) StorageS3Get(ctx context.Context, params api.StorageS3GetParams) (*api.StorageS3, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}

func (h *Handler) StorageS3Update(ctx context.Context, req *api.StorageS3, params api.StorageS3UpdateParams) (*api.StorageS3, error) {
	//TODO implement me
	return nil, ErrWithCode(http.StatusInternalServerError, E("not implemented"))
}
