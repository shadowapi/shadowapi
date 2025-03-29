package handler

import (
	"context"
	"encoding/json"
	"github.com/go-faster/jx"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// SyncpolicyCreate implements syncpolicy-create operation.
//
// Create a new sync policy.
//
// POST /syncpolicy
func (h *Handler) SyncpolicyCreate(ctx context.Context, req *api.SyncPolicy) (*api.SyncPolicy, error) {
	log := h.log.With("handler", "SyncpolicyCreate")

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.SyncPolicy, error) {
		policyUUID := uuid.Must(uuid.NewV7())

		// Convert userID from string -> *uuid.UUID (if userID is set)
		var userUUID *uuid.UUID
		if req.UserID != "" {
			uID, err := uuid.FromString(req.UserID)
			if err != nil {
				log.Error("invalid user ID", "error", err)
				return nil, ErrWithCode(http.StatusBadRequest, E("invalid user ID"))
			}
			userUUID = &uID
		}

		// Marshal the settings (if any) to JSON.
		var settingsData []byte
		if req.Settings.IsSet() {
			j, err := json.Marshal(req.Settings.Value)
			if err != nil {
				log.Error("failed to marshal syncpolicy settings", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal syncpolicy settings"))
			}
			settingsData = j
		}

		qParams := query.CreateSyncPolicyParams{
			UUID:        policyUUID,
			UserID:      pgtype.UUID{},
			Service:     req.Service,
			Blocklist:   req.Blocklist,
			ExcludeList: req.ExcludeList,
			SyncAll:     req.SyncAll,
			Settings:    settingsData,
		}

		// If userUUID is set, assign it to qParams
		if userUUID != nil {
			qParams.UserID = pgtype.UUID{
				Bytes: uToBytes(*userUUID),
				Valid: true,
			}
		}

		pol, err := query.New(tx).CreateSyncPolicy(ctx, qParams)
		if err != nil {
			log.Error("failed to create sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to create sync policy"))
		}

		out, err := qToApiSyncPolicy(pol)
		if err != nil {
			log.Error("failed to map sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map sync policy"))
		}

		return &out, nil
	})
}

// SyncpolicyDelete implements syncpolicy-delete operation.
//
// Delete a sync policy by uuid.
//
// DELETE /syncpolicy/{uuid}
func (h *Handler) SyncpolicyDelete(ctx context.Context, params api.SyncpolicyDeleteParams) error {
	log := h.log.With("handler", "SyncpolicyDelete")

	policyUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid policy uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid syncpolicy UUID"))
	}

	// Attempt deletion
	err = query.New(h.dbp).DeleteSyncPolicy(ctx, policyUUID)
	if err == pgx.ErrNoRows {
		// Possibly return 404 if desired, or treat it as success
		log.Warn("no sync policy found to delete", "policy_uuid", policyUUID)
		return nil
	} else if err != nil {
		log.Error("failed to delete sync policy", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete sync policy"))
	}

	return nil
}

// SyncpolicyGet implements syncpolicy-get operation.
//
// Retrieve a specific sync policy by uuid.
//
// GET /syncpolicy/{uuid}
func (h *Handler) SyncpolicyGet(ctx context.Context, params api.SyncpolicyGetParams) (*api.SyncPolicy, error) {
	log := h.log.With("handler", "SyncpolicyGet")

	policyUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid policy uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid syncpolicy UUID"))
	}

	// We'll do a single-get approach. We can reuse the "GetSyncPolicies" or custom logic.
	// We only have "GetSyncPolicies" (list) or "ListSyncPolicies" in queries, plus "delete" and "create" and "update".
	// We'll do "GetSyncPolicies" with a limit=1 and matching uuid. Or we can define a custom approach.

	policies, err := query.New(h.dbp).GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
		OrderBy:        nil,
		OrderDirection: "asc",
		Offset:         0,
		Limit:          1,
		Service:        "",
		UUID: pgtype.UUID{
			Bytes: uToBytes(policyUUID),
			Valid: true,
		},
		UserID:  pgtype.UUID{}, // not filtering by user
		SyncAll: false,         // not filtering
	})
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to get sync policy", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get sync policy"))
	}
	if len(policies) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("sync policy not found"))
	}

	out, err := qToApiSyncPolicyRow(policies[0])
	if err != nil {
		log.Error("failed to map sync policy row", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map sync policy row"))
	}

	return &out, nil
}

// SyncpolicyList implements syncpolicy-list operation.
//
// Retrieve a list of sync policies for the authenticated user.
//
// GET /syncpolicy
func (h *Handler) SyncpolicyList(ctx context.Context, params api.SyncpolicyListParams) (*api.SyncpolicyListOK, error) {
	log := h.log.With("handler", "SyncpolicyList")

	limit := int32(50)
	offset := int32(0)
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}
	if params.Offset.IsSet() {
		offset = params.Offset.Value
	}

	// We'll reuse the "GetSyncPolicies" which can do filtering.
	// For example, if we had user info in context, we could filter by user.
	qParams := query.GetSyncPoliciesParams{
		OrderBy:        nil,
		OrderDirection: "desc",
		Offset:         offset,
		Limit:          limit,
		Service:        "",            // or we can see if params has a service filter
		UUID:           pgtype.UUID{}, // no specific policy filter
		UserID:         pgtype.UUID{}, // might want to filter by user if known
		SyncAll:        false,         // no filter
	}

	pols, err := query.New(h.dbp).GetSyncPolicies(ctx, qParams)
	if err != nil && err != pgx.ErrNoRows {
		log.Error("failed to list sync policies", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list sync policies"))
	}

	out := &api.SyncpolicyListOK{}
	for _, row := range pols {
		apiItem, err := qToApiSyncPolicyRow(row)
		if err != nil {
			log.Error("failed to map sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map sync policy"))
		}
		out.Policies = append(out.Policies, apiItem)
	}

	return out, nil
}

// SyncpolicyUpdate implements syncpolicy-update operation.
//
// Update a sync policy by uuid.
//
// PUT /syncpolicy/{uuid}
func (h *Handler) SyncpolicyUpdate(ctx context.Context, req *api.SyncPolicy, params api.SyncpolicyUpdateParams) (*api.SyncPolicy, error) {
	log := h.log.With("handler", "SyncpolicyUpdate")

	policyUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid policy uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid syncpolicy UUID"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.SyncPolicy, error) {
		// see if it exists
		existing, err := query.New(tx).GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
			OrderBy:        nil,
			OrderDirection: "asc",
			Offset:         0,
			Limit:          1,
			Service:        "",
			UUID: pgtype.UUID{
				Bytes: uToBytes(policyUUID),
				Valid: true,
			},
			UserID:  pgtype.UUID{},
			SyncAll: false,
		})
		if err != nil && err != pgx.ErrNoRows {
			log.Error("failed to get sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get sync policy"))
		}
		if len(existing) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("sync policy not found"))
		}

		// Convert userID from string -> *uuid.UUID (if userID is set)
		var userUUID *uuid.UUID
		if req.UserID != "" {
			uID, err := uuid.FromString(req.UserID)
			if err != nil {
				log.Error("invalid user ID", "error", err)
				return nil, ErrWithCode(http.StatusBadRequest, E("invalid user ID"))
			}
			userUUID = &uID
		}

		// handle settings
		var settingsData []byte
		if req.Settings.IsSet() {
			j, err := json.Marshal(req.Settings.Value)
			if err != nil {
				log.Error("failed to marshal updated syncpolicy settings", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal updated syncpolicy settings"))
			}
			settingsData = j
		} else {
			// if not set, keep existing
			settingsData = existing[0].Settings
		}

		uParams := query.UpdateSyncPolicyParams{
			UserID: pgtype.UUID{},
			Service: func() string {
				if req.Service != "" {
					return req.Service
				}
				return existing[0].Service
			}(),
			Blocklist: func() []string {
				if req.Blocklist != nil {
					return req.Blocklist
				}
				return existing[0].Blocklist
			}(),
			ExcludeList: func() []string {
				if req.ExcludeList != nil {
					return req.ExcludeList
				}
				return existing[0].ExcludeList
			}(),
			SyncAll: func() bool {
				// If not set, fallback to existing
				return req.SyncAll
			}(),
			Settings: settingsData,
			UUID:     policyUUID,
		}

		if userUUID != nil {
			uParams.UserID = pgtype.UUID{
				Bytes: uToBytes(*userUUID),
				Valid: true,
			}
		} else {
			// keep existing
			if existing[0].UserID.Valid {
				uParams.UserID = existing[0].UserID
			}
		}

		err = query.New(tx).UpdateSyncPolicy(ctx, uParams)
		if err == pgx.ErrNoRows {
			return nil, ErrWithCode(http.StatusNotFound, E("sync policy not found"))
		} else if err != nil {
			log.Error("failed to update sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update sync policy"))
		}

		// re-fetch
		outP, err := query.New(tx).GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
			OrderBy:        nil,
			OrderDirection: "asc",
			Offset:         0,
			Limit:          1,
			Service:        "",
			UUID: pgtype.UUID{
				Bytes: uToBytes(policyUUID),
				Valid: true,
			},
			UserID:  pgtype.UUID{},
			SyncAll: false,
		})
		if err != nil && err != pgx.ErrNoRows {
			log.Error("failed to get updated sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get updated sync policy"))
		}
		if len(outP) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("sync policy not found after update"))
		}

		final, err := qToApiSyncPolicyRow(outP[0])
		if err != nil {
			log.Error("failed to map sync policy row after update", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map sync policy row"))
		}

		return &final, nil
	})
}

/* ------------------------------------------------------------------
   Helpers below
------------------------------------------------------------------ */

func qToApiSyncPolicy(dbp query.SyncPolicy) (api.SyncPolicy, error) {
	out := api.SyncPolicy{
		UUID:        dbp.UUID.String(),
		Service:     dbp.Service,
		Blocklist:   dbp.Blocklist,
		ExcludeList: dbp.ExcludeList,
		SyncAll:     dbp.SyncAll,
		CreatedAt:   dbp.CreatedAt.Time,
		UpdatedAt:   dbp.UpdatedAt.Time,
	}
	if dbp.UserID.Valid {
		// user id is stored
		uid, err := bytesToUUID(dbp.UserID.Bytes)
		if err == nil {
			out.UserID = uid.String()
		}
	}
	// parse settings if any
	if len(dbp.Settings) > 0 {

		// Suppose you unmarshalled into:
		var mp map[string]json.RawMessage

		// Convert it to SyncPolicySettings:
		spSettings := make(api.SyncPolicySettings)
		for k, v := range mp {
			// v is json.RawMessage, cast it to jx.Raw
			spSettings[k] = jx.Raw(v)
		}

		// Then assign:
		out.Settings.Set = true
		out.Settings.Value = spSettings

	}

	return out, nil
}

func qToApiSyncPolicyRow(row query.GetSyncPoliciesRow) (api.SyncPolicy, error) {
	var sp query.SyncPolicy
	sp.UUID = row.UUID
	sp.UserID = row.UserID
	sp.Service = row.Service
	sp.Blocklist = row.Blocklist
	sp.ExcludeList = row.ExcludeList
	sp.SyncAll = row.SyncAll
	sp.Settings = row.Settings
	sp.CreatedAt = row.CreatedAt
	sp.UpdatedAt = row.UpdatedAt

	return qToApiSyncPolicy(sp)
}

func uToBytes(u uuid.UUID) [16]byte {
	var arr [16]byte
	copy(arr[:], u.Bytes())
	return arr
}

func bytesToUUID(b [16]byte) (uuid.UUID, error) {
	return uuid.FromBytes(b[:])
}
