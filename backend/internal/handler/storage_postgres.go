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
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) StoragePostgresCreate(ctx context.Context, req *api.StoragePostgres) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresCreate")

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
		UUID:      pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
		Name:      req.Name,
		Type:      "postgres",
		IsEnabled: isEnabled,
		Settings:  settings,
	})
	if err != nil {
		log.Error("failed to create storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create storage"))
	}

	resp := *req
	resp.UUID = api.OptString{Value: storage.UUID.String(), Set: true}

	return &resp, nil
}

func (h *Handler) StoragePostgresDelete(ctx context.Context, params api.StoragePostgresDeleteParams) error {
	log := h.log.With("handler", "StoragePostgresDelete")

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	if err := query.New(h.dbp).DeleteStorage(ctx, pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true}); err != nil {
		log.Error("failed to delete storage", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete storage"))
	}

	return nil
}

func (h *Handler) StoragePostgresGet(ctx context.Context, params api.StoragePostgresGetParams) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresGet")

	id, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}
	fmt.Println("id", id)

	storages, err := query.New(h.dbp).GetStorage(ctx, pgtype.UUID{Bytes: converter.UToBytes(id), Valid: true})
	if err != nil {
		log.Error("failed to get storage", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get storage"))
	}

	return QToStoragePostgres(storages)
}

func (h *Handler) StoragePostgresUpdate(ctx context.Context, req *api.StoragePostgres, params api.StoragePostgresUpdateParams) (*api.StoragePostgres, error) {
	log := h.log.With("handler", "StoragePostgresUpdate")

	// Validate table definitions
	if err := validatePostgresTables(req.Tables); err != nil {
		return nil, err
	}

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.StoragePostgres, error) {
		storage, err := query.New(tx).GetStorage(ctx, pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true})
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

		if err := query.New(h.dbp).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:      pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			Type:      "postgres",
			Name:      req.Name,
			IsEnabled: isEnabled,
			Settings:  newSettings,
		}); err != nil {
			log.Error("failed to update storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		return h.StoragePostgresGet(ctx, api.StoragePostgresGetParams{UUID: params.UUID})
	})
}

func QToStoragePostgres(row query.GetStorageRow) (*api.StoragePostgres, error) {
	var s api.StoragePostgres
	if err := json.Unmarshal(row.Storage.Settings, &s); err != nil {
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to unmarshal postgres settings: %w", err))
	}
	s.UUID = api.NewOptString(row.Storage.UUID.String())
	s.Name = row.Storage.Name
	s.IsEnabled = api.NewOptBool(row.Storage.IsEnabled)
	return &s, nil
}

func (h *Handler) StoragePostgresTablesReplace(ctx context.Context, req []api.StoragePostgresTable, params api.StoragePostgresTablesReplaceParams) ([]api.StoragePostgresTable, error) {
	log := h.log.With("handler", "StoragePostgresTablesReplace")

	// Validate table definitions
	if err := validatePostgresTables(req); err != nil {
		return nil, err
	}

	storageUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse storage uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid storage UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.StoragePostgresTable, error) {
		// Get existing storage
		storage, err := query.New(tx).GetStorage(ctx, pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true})
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
		if err := query.New(tx).UpdateStorage(ctx, query.UpdateStorageParams{
			UUID:      pgtype.UUID{Bytes: converter.UToBytes(storageUUID), Valid: true},
			Type:      storage.Storage.Type,
			Name:      storage.Storage.Name,
			IsEnabled: storage.Storage.IsEnabled,
			Settings:  newSettings,
		}); err != nil {
			log.Error("failed to update storage", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update storage"))
		}

		return req, nil
	})
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
