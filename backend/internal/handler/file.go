package handler

import (
	"context"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

func (h *Handler) GenerateDownloadLink(ctx context.Context, req *api.GenerateDownloadLinkRequest) (*api.GenerateDownloadLinkResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) GeneratePresignedUploadUrl(ctx context.Context, req *api.UploadPresignedUrlRequest) (*api.UploadPresignedUrlResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) UploadFile(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	//TODO implement me
	panic("implement me")
}
