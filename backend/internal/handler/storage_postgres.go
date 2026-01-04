// storage_postgres.go
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/converter"
	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/workspace"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StoragePostgresCreate(ctx context.Context, req *api.StoragePostgres) (api.StoragePostgresCreateRes, error) {
	log := h.log.With("handler", "StoragePostgresCreate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Validate table definitions
	if err := validatePostgresTables(req.Tables); err != nil {
		return nil, err
	}

	storageUUID := uuid.Must(uuid.NewV7())
	settings, err := json.Marshal(req)
	if err != nil {
		log.Error("failed to marshal settings", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
	}

	// Extract underlying values from optional fields.
	var isEnabled bool
	if req.IsEnabled.IsSet() {
		isEnabled = req.IsEnabled.Value
	} else {
		isEnabled = false
	}

	storage, err := query.New(h.dbp).CreateStorage(ctx, query.CreateStorageParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
		Name:          req.Name,
		Type:          "postgres",
		IsEnabled:     isEnabled,
		Settings:      settings,
	})
	if err != nil {
		log.Error("failed to create storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create storage"))
	}

	resp := *req
	resp.UUID = api.OptString{Value: storage.UUID.String(), Set: true}

	return &resp, nil
}

func (h *Handler) StoragePostgresDelete(ctx context.Context, params api.StoragePostgresDeleteParams) (api.StoragePostgresDeleteRes, error) {
	log := h.log.With("handler", "StoragePostgresDelete")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorageByWorkspace(ctx, query.DeleteStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
		WorkspaceUUID: workspaceUUID,
	}); err != nil {
		log.Error("failed to delete storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to delete storage"))
	}

	return &api.StoragePostgresDeleteOK{}, nil
}

func (h *Handler) StoragePostgresGet(ctx context.Context, params api.StoragePostgresGetParams) (api.StoragePostgresGetRes, error) {
	log := h.log.With("handler", "StoragePostgresGet")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	fmt.Println("id", id)

	storages, err := query.New(h.dbp).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(id), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	return QToStoragePostgresByWorkspace(storages)
}

func (h *Handler) StoragePostgresUpdate(ctx context.Context, req *api.StoragePostgres, params api.StoragePostgresUpdateParams) (api.StoragePostgresUpdateRes, error) {
	log := h.log.With("handler", "StoragePostgresUpdate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Validate table definitions
	if err := validatePostgresTables(req.Tables); err != nil {
		return nil, err
	}

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.StoragePostgresUpdateRes, error) {
		storage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}

		// For update, use new values if provided; otherwise fall back to current DB values.
		var isEnabled bool
		if req.IsEnabled.IsSet() {
			isEnabled = req.IsEnabled.Value
		} else {
			isEnabled = storage.Storage.IsEnabled
		}

		newSettings, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}

		if err := query.New(tx).UpdateStorageByWorkspace(ctx, query.UpdateStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
			Type:          "postgres",
			Name:          req.Name,
			IsEnabled:     isEnabled,
			Settings:      newSettings,
		}); err != nil {
			log.Error("failed to update storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		// Re-fetch and return the updated storage
		updatedStorage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			log.Error("failed to get updated storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated storage"))
		}
		return QToStoragePostgresByWorkspace(updatedStorage)
	})
}

func QToStoragePostgres(row query.GetStorageRow) (*api.StoragePostgres, error) {
	return storageToPostgres(row.Storage)
}

func QToStoragePostgresByWorkspace(row query.GetStorageByWorkspaceRow) (*api.StoragePostgres, error) {
	return storageToPostgres(row.Storage)
}

func storageToPostgres(s query.Storage) (*api.StoragePostgres, error) {
	var out api.StoragePostgres
	if err := json.Unmarshal(s.Settings, &out); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal postgres settings: %w", err))
	}
	out.UUID = api.NewOptString(s.UUID.String())
	out.Name = s.Name
	out.IsEnabled = api.NewOptBool(s.IsEnabled)
	return &out, nil
}

func (h *Handler) StoragePostgresTablesReplace(ctx context.Context, req []api.StoragePostgresTable, params api.StoragePostgresTablesReplaceParams) (api.StoragePostgresTablesReplaceRes, error) {
	log := h.log.With("handler", "StoragePostgresTablesReplace")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Validate table definitions
	if err := validatePostgresTables(req); err != nil {
		return nil, err
	}

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (api.StoragePostgresTablesReplaceRes, error) {
		// Get existing storage
		storage, err := query.New(tx).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
			}
			log.Error("failed to get storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
		}

		// Verify storage is postgres type
		if storage.Storage.Type != "postgres" {
			return nil, ErrWithCode(http.StatusBadRequest, E("storage is not of type postgres"))
		}

		// Unmarshal existing settings as raw JSON map to preserve all fields
		var existingSettings map[string]json.RawMessage
		if err := json.Unmarshal(storage.Storage.Settings, &existingSettings); err != nil {
			log.Error("failed to unmarshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal settings"))
		}

		// Marshal tables to JSON and replace in settings
		tablesJSON, err := json.Marshal(req)
		if err != nil {
			log.Error("failed to marshal tables", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal tables"))
		}
		existingSettings["tables"] = tablesJSON

		// Marshal updated settings
		newSettings, err := json.Marshal(existingSettings)
		if err != nil {
			log.Error("failed to marshal settings", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal settings"))
		}

		// Update storage with new settings (preserve all other fields)
		if err := query.New(tx).UpdateStorageByWorkspace(ctx, query.UpdateStorageByWorkspaceParams{
			UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			WorkspaceUUID: workspaceUUID,
			Type:          storage.Storage.Type,
			Name:          storage.Storage.Name,
			IsEnabled:     storage.Storage.IsEnabled,
			Settings:      newSettings,
		}); err != nil {
			log.Error("failed to update storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		res := api.StoragePostgresTablesReplaceOKApplicationJSON(req)
		return &res, nil
	})
}

// StoragePostgresIntrospectTables lists all tables in the connected PostgreSQL database.
func (h *Handler) StoragePostgresIntrospectTables(ctx context.Context, params api.StoragePostgresIntrospectTablesParams) (api.StoragePostgresIntrospectTablesRes, error) {
	log := h.log.With("handler", "StoragePostgresIntrospectTables")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Validate storage belongs to workspace before introspection
	storageUUIDParsed, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	_, err = query.New(h.dbp).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUIDParsed), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
	}

	tables, err := h.storageManager.ListTables(ctx, params.UUID)
	if err != nil {
		log.Error("failed to list tables", "error", err, "storageUUID", params.UUID)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list tables: %s", err.Error()))
	}

	result := make([]api.StoragePostgresIntrospectTable, len(tables))
	for i, t := range tables {
		result[i] = api.StoragePostgresIntrospectTable{
			Name:          t.Name,
			RowCount:      api.NewOptInt(int(t.RowCount)),
			HasPrimaryKey: api.NewOptBool(t.HasPrimaryKey),
		}
	}

	return &api.StoragePostgresIntrospectTablesResponse{
		Tables: result,
	}, nil
}

// StoragePostgresIntrospectTable returns the schema information for a specific table.
func (h *Handler) StoragePostgresIntrospectTable(ctx context.Context, params api.StoragePostgresIntrospectTableParams) (api.StoragePostgresIntrospectTableRes, error) {
	log := h.log.With("handler", "StoragePostgresIntrospectTable")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Validate storage belongs to workspace before introspection
	storageUUIDParsed, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	_, err = query.New(h.dbp).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUIDParsed), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
	}

	schema, err := h.storageManager.GetTableFields(ctx, params.UUID, params.TableName)
	if err != nil {
		log.Error("failed to get table fields", "error", err, "storageUUID", params.UUID, "tableName", params.TableName)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get table fields: %s", err.Error()))
	}

	fields := make([]api.StoragePostgresIntrospectField, len(schema.Fields))
	for i, f := range schema.Fields {
		fields[i] = api.StoragePostgresIntrospectField{
			Name:         f.Name,
			Type:         api.StoragePostgresIntrospectFieldType(f.Type),
			PgType:       api.NewOptString(f.PgType),
			Nullable:     f.Nullable,
			IsPrimaryKey: f.IsPrimaryKey,
			DefaultValue: api.NewOptString(f.DefaultValue),
		}
	}

	return &api.StoragePostgresIntrospectTableResponse{
		Name:     schema.Name,
		Exists:   schema.Exists,
		Fields:   fields,
		RowCount: api.NewOptInt(int(schema.RowCount)),
	}, nil
}

// StoragePostgresTablesCreate creates a new table in the PostgreSQL database.
func (h *Handler) StoragePostgresTablesCreate(ctx context.Context, req *api.StoragePostgresTableCreateRequest, params api.StoragePostgresTablesCreateParams) (api.StoragePostgresTablesCreateRes, error) {
	log := h.log.With("handler", "StoragePostgresTablesCreate")

	// Extract workspace UUID from context - required for workspace-scoped access
	workspaceUUID, err := workspace.RequireWorkspaceUUID(ctx)
	if err != nil {
		log.Error("workspace context required", "error", err)
		return nil, ErrWithCode(http.StatusUnauthorized, E("workspace context required"))
	}

	// Validate storage belongs to workspace before creating table
	storageUUIDParsed, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	_, err = query.New(h.dbp).GetStorageByWorkspace(ctx, query.GetStorageByWorkspaceParams{
		UUID:          pgtype.UUID{Bytes: converter.UToBytes(storageUUIDParsed), Valid: true},
		WorkspaceUUID: workspaceUUID,
	})
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusNotFound, E("storage not found"))
	}

	dropIfExists := false
	if req.DropIfExists.IsSet() {
		dropIfExists = req.DropIfExists.Value
	}

	wasDropped, err := h.storageManager.CreateTable(ctx, params.UUID, req.Name, req.Fields, dropIfExists)
	if err != nil {
		log.Error("failed to create table", "error", err, "storageUUID", params.UUID, "tableName", req.Name)
		return &api.StoragePostgresTableCreateResponse{
			Success:   false,
			TableName: api.NewOptString(req.Name),
			Error:     api.NewOptString(err.Error()),
		}, nil
	}

	return &api.StoragePostgresTableCreateResponse{
		Success:    true,
		TableName:  api.NewOptString(req.Name),
		WasDropped: api.NewOptBool(wasDropped),
	}, nil
}

// validatePostgresTables validates table definitions for a PostgreSQL storage.
func validatePostgresTables(tables []api.StoragePostgresTable) error {
	if len(tables) == 0 {
		return nil // Tables are optional
	}

	tableNames := make(map[string]bool)
	for _, table := range tables {
		// Check unique table names
		if tableNames[table.Name] {
			return ErrWithCode(http.StatusBadRequest, E("duplicate table name: %s", table.Name))
		}
		tableNames[table.Name] = true

		// Validate fields
		if len(table.Fields) == 0 {
			return ErrWithCode(http.StatusBadRequest, E("table %s must have at least one field", table.Name))
		}

		fieldNames := make(map[string]bool)
		pkCount := 0
		for _, field := range table.Fields {
			// Check unique field names within table
			if fieldNames[field.Name] {
				return ErrWithCode(http.StatusBadRequest, E("duplicate field name in table %s: %s", table.Name, field.Name))
			}
			fieldNames[field.Name] = true

			// Count primary keys
			if field.IsPrimaryKey.Value {
				pkCount++
			}

			// Primary key cannot be nullable
			if field.IsPrimaryKey.Value && field.Nullable.Value {
				return ErrWithCode(http.StatusBadRequest, E("primary key field %s in table %s cannot be nullable", field.Name, table.Name))
			}
		}

		// At most one primary key per table
		if pkCount > 1 {
			return ErrWithCode(http.StatusBadRequest, E("table %s has multiple primary keys (%d)", table.Name, pkCount))
		}
	}
	return nil
}
