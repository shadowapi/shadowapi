package handler

import (
	"context"
	ht "github.com/ogen-go/ogen/http"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// FileCreate implements file-create operation.
//
// Upload a new file and create its record.
//
// POST /file
func (h *Handler) FileCreate(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	var r *api.UploadFileResponse
	return r, ht.ErrNotImplemented
}

// FileDelete implements file-delete operation.
//
// Delete a stored file.
//
// DELETE /file/{uuid}
func (h *Handler) FileDelete(ctx context.Context, params api.FileDeleteParams) error {
	return ht.ErrNotImplemented
}

// FileGet implements file-get operation.
//
// Retrieve details of a stored file.
//
// GET /file/{uuid}
func (h *Handler) FileGet(ctx context.Context, params api.FileGetParams) (*api.FileObject, error) {
	var r *api.FileObject
	return r, ht.ErrNotImplemented
}

// FileList implements file-list operation.
//
// Retrieve a list of stored files.
//
// GET /file
func (h *Handler) FileList(ctx context.Context, params api.FileListParams) ([]api.FileObject, error) {
	var r []api.FileObject
	return r, ht.ErrNotImplemented
}

// FileUpdate implements file-update operation.
//
// Update metadata of a stored file.
//
// PUT /file/{uuid}
func (h *Handler) FileUpdate(ctx context.Context, req *api.FileUpdateReq, params api.FileUpdateParams) (*api.FileObject, error) {
	var r *api.FileObject
	return r, ht.ErrNotImplemented
}

func (h *Handler) GenerateDownloadLink(ctx context.Context, req *api.GenerateDownloadLinkRequest) (*api.GenerateDownloadLinkResponse, error) {
	//TODO implement me
	var pl *api.GenerateDownloadLinkResponse
	return pl, ht.ErrNotImplemented
}

func (h *Handler) GeneratePresignedUploadUrl(ctx context.Context, req *api.UploadPresignedUrlRequest) (*api.UploadPresignedUrlResponse, error) {
	//TODO implement me
	var pl *api.UploadPresignedUrlResponse
	return pl, ht.ErrNotImplemented
}

func (h *Handler) UploadFile(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	//TODO implement me
	var pl *api.UploadFileResponse
	return pl, ht.ErrNotImplemented
}
