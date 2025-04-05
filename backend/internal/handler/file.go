package handler

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// TODO @reactima see possible issues!
// DEDUBLICATE!
// 1. FileCreate implements file-create operation.
// 2. Upload a new file and create its record.
// POST /file
//
// Possible issues or limitations:
//  1. No actual file binary data is stored (just metadata).
//  2. We require a non-empty StorageUUID and validate it against a storage table row.
//  3. The fallback storage type is "hostfiles" or could come from the storage row details—
//     you might adapt this to read from the storage row, e.g. row.Type = "s3"/"postgres" etc.
//  4. We default the file name to "untitled_file" if none is provided.
func (h *Handler) FileCreate(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	log := h.log.With("handler", "FileCreate")

	// 1. Ensure storage_uuid is present
	if req.StorageUUID == "" {
		log.Error("no storage_uuid provided in request")
		return nil, ErrWithCode(http.StatusBadRequest, E("storage_uuid is required"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.UploadFileResponse, error) {
		// 1. Validate StorageUUID and fetch the row
		sUUID, err := uuid.FromString(req.StorageUUID)
		if err != nil {
			log.Error("invalid storage_uuid", "error", err)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
		}
		storageRow, err := query.New(tx).GetStorage(ctx, pgtype.UUID{Bytes: uToBytes(sUUID), Valid: true})
		if err == pgx.ErrNoRows {
			log.Error("storage row not found", "uuid", req.StorageUUID)
			return nil, ErrWithCode(http.StatusBadRequest, E("storage not found"))
		} else if err != nil {
			log.Error("failed to get storage row", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage row"))
		}

		// 2. Generate file UUID and gather other fields
		fileUUID := uuid.Must(uuid.NewV7())
		name := req.Name.Or("untitled_file")
		mimeType := req.MimeType.Or("application/octet-stream")

		// 3. Insert into DB using the type from the storage row
		fileRow, err := query.New(tx).CreateFile(ctx, query.CreateFileParams{
			UUID:        pgtype.UUID{Bytes: uToBytes(fileUUID), Valid: true},
			StorageUuid: pgtype.UUID{Bytes: uToBytes(sUUID), Valid: true},
			StorageType: storageRow.Storage.Type,
			Name:        name,
			MimeType: pgtype.Text{
				String: mimeType,
				Valid:  true,
			},
			Size: pgtype.Int8{
				Int64: 0,
				Valid: true,
			},
		})
		if err != nil {
			log.Error("failed to create file record", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create file record"))
		}

		// 4. Return the newly created record
		resp := &api.UploadFileResponse{
			File: api.NewOptFileObject(api.FileObject{
				UUID:        api.NewOptString(fileRow.UUID.String()),
				StorageUUID: req.StorageUUID,
				StorageType: storageRow.Storage.Type,
				Name:        fileRow.Name,
				MimeType:    api.NewOptString(fileRow.MimeType.String),
				Size:        api.NewOptInt(int(fileRow.Size.Int64)),
				CreatedAt:   api.NewOptDateTime(fileRow.CreatedAt.Time),
				UpdatedAt:   api.NewOptDateTime(fileRow.UpdatedAt.Time),
			}),
		}

		return resp, nil
	})
}

// FileDelete implements file-delete operation.
// Delete a stored file.
// DELETE /file/{uuid}
//
// Possible issues or limitations:
// 1. Returning 200 if file doesn't exist (warn log). You might want to return 404 instead.
// 2. We do no cleanup of actual file data if it exists on disk or S3; it's purely a metadata delete.
func (h *Handler) FileDelete(ctx context.Context, params api.FileDeleteParams) error {
	log := h.log.With("handler", "FileDelete")

	fileUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid file UUID", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid file UUID"))
	}

	if err := query.New(h.dbp).DeleteFile(ctx, pgtype.UUID{Bytes: [16]byte(fileUUID.Bytes()), Valid: true}); err == pgx.ErrNoRows {
		log.Warn("no file found to delete", "file_uuid", fileUUID)
		return nil // or return 404 if you prefer to surface that?
	} else if err != nil {
		log.Error("failed to delete file", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete file"))
	}

	return nil
}

// FileGet implements file-get operation.
// Retrieve details of a stored file.
// GET /file/{uuid}
//
// Possible issues or limitations:
// 1. We only fetch metadata, not the actual file content.
// 2. The “StorageUUID” field is blank if we haven't associated a storage row.
func (h *Handler) FileGet(ctx context.Context, params api.FileGetParams) (*api.FileObject, error) {
	log := h.log.With("handler", "FileGet")

	fileUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid file UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid file UUID"))
	}

	fileRow, err := query.New(h.dbp).GetFile(ctx, pgtype.UUID{Bytes: uToBytes(fileUUID), Valid: true})
	if err == pgx.ErrNoRows {
		return nil, ErrWithCode(http.StatusNotFound, E("file not found"))
	} else if err != nil {
		log.Error("failed to get file", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get file"))
	}

	out := &api.FileObject{
		UUID:        api.NewOptString(fileRow.File.UUID.String()),
		Name:        fileRow.File.Name,
		MimeType:    api.NewOptString(fileRow.File.MimeType.String),
		Size:        api.NewOptInt(int(fileRow.File.Size.Int64)),
		StorageType: "", // Not populated here
		StorageUUID: "", // Not populated
	}
	if fileRow.File.CreatedAt.Valid {
		out.CreatedAt = api.NewOptDateTime(fileRow.File.CreatedAt.Time)
	}
	if fileRow.File.UpdatedAt.Valid {
		out.UpdatedAt = api.NewOptDateTime(fileRow.File.UpdatedAt.Time)
	}

	return out, nil
}

// FileList implements file-list operation.
// Retrieve a list of stored files.
// GET /file
//
// Possible issues or limitations:
// 1. Very simple pagination logic (limit + offset).
// 2. We don't return total count, so the client doesn't know how many more exist.
func (h *Handler) FileList(ctx context.Context, params api.FileListParams) ([]api.FileObject, error) {
	log := h.log.With("handler", "FileList")

	limit := int32(50)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}

	files, err := query.New(h.dbp).ListFiles(ctx, query.ListFilesParams{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to list files", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list files"))
	}

	var results []api.FileObject
	for _, f := range files {
		fo := api.FileObject{
			UUID:        api.NewOptString(f.File.UUID.String()),
			Name:        f.File.Name,
			MimeType:    api.NewOptString(f.File.MimeType.String),
			Size:        api.NewOptInt(int(f.File.Size.Int64)),
			StorageType: f.File.StorageType,
			StorageUUID: "",
		}
		if f.File.CreatedAt.Valid {
			fo.CreatedAt = api.NewOptDateTime(f.File.CreatedAt.Time)
		}
		if f.File.UpdatedAt.Valid {
			fo.UpdatedAt = api.NewOptDateTime(f.File.UpdatedAt.Time)
		}
		results = append(results, fo)
	}

	return results, nil
}

// FileUpdate implements file-update operation.
// Update metadata of a stored file.
// PUT /file/{uuid}
//
// This example only updates the file's Name field for brevity.
// We keep the old StorageType, StorageUuid, MimeType, Size, Data, etc.
// If you need to update other fields, expand the code accordingly.
func (h *Handler) FileUpdate(ctx context.Context, req *api.FileUpdateReq, params api.FileUpdateParams) (*api.FileObject, error) {
	log := h.log.With("handler", "FileUpdate")

	fileUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid file UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid file UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.FileObject, error) {
		// 1. Fetch the record to ensure it exists
		oldFileRow, err := query.New(tx).GetFile(ctx, pgtype.UUID{Bytes: uToBytes(fileUUID), Valid: true})
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("file not found"))
		} else if err != nil {
			log.Error("failed to fetch file for update", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get file record"))
		}

		// 2. Update it, preserving the existing fields not being changed
		err = query.New(tx).UpdateFile(ctx, query.UpdateFileParams{
			StorageType: oldFileRow.File.StorageType,
			StorageUuid: pgtype.UUID{Bytes: [16]byte(oldFileRow.File.StorageUuid.Bytes()), Valid: true},
			Name:        req.Name, // updated name from the request
			MimeType:    oldFileRow.File.MimeType,
			Size:        oldFileRow.File.Size,
			Data:        oldFileRow.File.Data,
			Path:        oldFileRow.File.Path,
			IsRaw:       oldFileRow.File.IsRaw,
			UUID:        pgtype.UUID{Bytes: uToBytes(fileUUID), Valid: true},
		})
		if err != nil {
			log.Error("failed to update file", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update file record"))
		}

		// 3. Fetch the updated record
		updatedRow, err := query.New(tx).GetFile(ctx, pgtype.UUID{Bytes: uToBytes(fileUUID), Valid: true})
		if err != nil {
			log.Error("failed to fetch file post-update", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to retrieve updated record"))
		}

		fo := &api.FileObject{
			UUID:        api.NewOptString(updatedRow.File.UUID.String()),
			Name:        updatedRow.File.Name,
			MimeType:    api.NewOptString(updatedRow.File.MimeType.String),
			Size:        api.NewOptInt(int(updatedRow.File.Size.Int64)),
			StorageType: updatedRow.File.StorageType,
			StorageUUID: "", // you can set updatedRow.File.StorageUuid if needed
		}
		if updatedRow.File.CreatedAt.Valid {
			fo.CreatedAt = api.NewOptDateTime(updatedRow.File.CreatedAt.Time)
		}
		if updatedRow.File.UpdatedAt.Valid {
			fo.UpdatedAt = api.NewOptDateTime(updatedRow.File.UpdatedAt.Time)
		}

		return fo, nil
	})
}

// GenerateDownloadLink implements a minimal "signed link" logic or placeholder.
//
// POST /file/download/link
//
// Possible issues or limitations:
// 1. We don't do real signing; we just append file UUID to a dummy URL.
// 2. We aren't verifying if the user has permissions to download the file.
func (h *Handler) GenerateDownloadLink(ctx context.Context, req *api.GenerateDownloadLinkRequest) (*api.GenerateDownloadLinkResponse, error) {
	log := h.log.With("handler", "GenerateDownloadLink")

	// Check that the file exists
	if !req.FileUUID.IsSet() {
		return nil, ErrWithCode(http.StatusBadRequest, E("file_uuid is required"))
	}
	_, err := uuid.FromString(req.FileUUID.Value)
	if err != nil {
		log.Error("invalid file UUID", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid file UUID"))
	}

	// Just do a placeholder link
	resp := &api.GenerateDownloadLinkResponse{
		URL: api.NewOptString("https://example.com/fake-signed-url?file=" + req.FileUUID.Value),
	}

	return resp, nil
}

// GeneratePresignedUploadUrl implements a typical presigned URL approach (placeholder).
//
// POST /file/upload/url
//
// Possible issues or limitations:
// 1. We do not store a DB record here automatically unless you adapt it.
// 2. The returned URL is just a dummy example.com endpoint.

// UploadFile is a placeholder for receiving actual file data in multipart/form-data.
//
// POST /file/upload
//
// Currently, we're only doing the same thing as FileCreate (metadata only):
// 1) Require storage_uuid.
// 2) Look up storage row in DB.
// 3) Insert file record with name & mimeType from request, or fallback defaults.
// 4) Return the newly created record.
func (h *Handler) UploadFile(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	log := h.log.With("handler", "UploadFile")

	if req.StorageUUID == "" {
		log.Error("no storage_uuid provided in request")
		return nil, ErrWithCode(http.StatusBadRequest, E("storage_uuid is required"))
	}

	resp, err := db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.UploadFileResponse, error) {
		sUUID, convErr := uuid.FromString(req.StorageUUID)
		if convErr != nil {
			log.Error("invalid storage_uuid", "error", convErr)
			return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
		}

		storageRow, queryErr := query.New(tx).GetStorage(ctx, pgtype.UUID{Bytes: uToBytes(sUUID), Valid: true})
		if queryErr == pgx.ErrNoRows {
			log.Error("storage row not found", "uuid", req.StorageUUID)
			return nil, ErrWithCode(http.StatusBadRequest, E("storage not found"))
		} else if queryErr != nil {
			log.Error("failed to get storage row", "error", queryErr)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage row"))
		}

		fileUUID := uuid.Must(uuid.NewV7())
		name := req.Name.Or("untitled_file")
		mimeType := req.MimeType.Or("application/octet-stream")

		fileRow, createErr := query.New(tx).CreateFile(ctx, query.CreateFileParams{
			UUID:        pgtype.UUID{Bytes: uToBytes(fileUUID), Valid: true},
			StorageType: storageRow.Storage.Type,
			StorageUuid: pgtype.UUID{}, // keep existing?
			Name:        name,
			MimeType: pgtype.Text{
				String: mimeType,
				Valid:  true,
			},
			Size: pgtype.Int8{
				Int64: 0,
				Valid: true,
			},
		})
		if createErr != nil {
			log.Error("failed to create file record", "error", createErr)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create file record"))
		}

		out := &api.UploadFileResponse{
			File: api.NewOptFileObject(api.FileObject{
				UUID:        api.NewOptString(fileRow.UUID.String()),
				StorageUUID: req.StorageUUID,
				StorageType: fileRow.StorageType,
				Name:        fileRow.Name,
				MimeType:    api.NewOptString(fileRow.MimeType.String),
				Size:        api.NewOptInt(int(fileRow.Size.Int64)),
				CreatedAt:   api.NewOptDateTime(fileRow.CreatedAt.Time),
				UpdatedAt:   api.NewOptDateTime(fileRow.UpdatedAt.Time),
			}),
		}
		return out, nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// nowUTC is a simple helper returning current time in UTC.
func nowUTC() time.Time {
	return time.Now().UTC()
}
