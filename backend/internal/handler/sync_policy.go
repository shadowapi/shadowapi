package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-faster/jx"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) SyncpolicyCreate(ctx context.Context, req *api.SyncPolicy) (*api.SyncPolicy, error) {
	log := h.log.With("handler", "SyncpolicyCreate")
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.SyncPolicy, error) {
		policyUUID := uuid.Must(uuid.NewV7())
		var userUUID interface{}
		if req.UserUUID.IsSet() && req.UserUUID.Value != "" {
			uID, err := uuid.FromString(req.UserUUID.Value)
			if err != nil {
				log.Error("invalid user uuid", "error", err)
				return nil, ErrWithCode(http.StatusBadRequest, E("invalid user uuid"))
			}
			userUUID = uID.String()
		}
		var settingsData []byte
		if req.Settings.IsSet() {
			j, err := json.Marshal(req.Settings.Value)
			if err != nil {
				log.Error("failed to marshal sync policy settings", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal sync policy settings"))
			}
			settingsData = j
		}
		qParams := query.CreateSyncPolicyParams{
			UUID:        pgtype.UUID{Bytes: uToBytes(policyUUID), Valid: true},
			UserUUID:    userUUID,
			Service:     req.Service,
			Blocklist:   req.Blocklist,
			ExcludeList: req.ExcludeList,
			SyncAll:     req.SyncAll.Or(false),
			Settings:    settingsData,
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

func (h *Handler) SyncpolicyDelete(ctx context.Context, params api.SyncpolicyDeleteParams) error {
	log := h.log.With("handler", "SyncpolicyDelete")
	policyUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid policy uuid", "error", err)
		return ErrWithCode(http.StatusBadRequest, E("invalid sync policy uuid"))
	}
	err = query.New(h.dbp).DeleteSyncPolicy(ctx, pgtype.UUID{Bytes: uToBytes(policyUUID), Valid: true})
	if err != nil {
		log.Error("failed to delete sync policy", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete sync policy"))
	}
	return nil
}

func (h *Handler) SyncpolicyGet(ctx context.Context, params api.SyncpolicyGetParams) (*api.SyncPolicy, error) {
	log := h.log.With("handler", "SyncpolicyGet")
	policyUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid policy uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid sync policy uuid"))
	}
	policies, err := query.New(h.dbp).GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
		OrderBy:        nil,
		OrderDirection: "asc",
		Offset:         0,
		Limit:          1,
		Service:        "",
		UUID:           pgtype.UUID{Bytes: uToBytes(policyUUID), Valid: true},
		UserUUID:       "",
		SyncAll:        -1,
	})
	if err != nil {
		log.Error("failed to get sync policy", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get sync policy"))
	}
	if len(policies) == 0 {
		return nil, ErrWithCode(http.StatusNotFound, E("sync policy not found"))
	}
	out, err := qToApiSyncPolicyRow(policies[0])
	if err != nil {
		log.Error("failed to map sync policy", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map sync policy"))
	}
	return &out, nil
}

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
	qParams := query.GetSyncPoliciesParams{
		OrderBy:        nil,
		OrderDirection: "desc",
		Offset:         offset,
		Limit:          limit,
		Service:        "",
		UUID:           "",
		UserUUID:       "",
		SyncAll:        -1,
	}
	pols, err := query.New(h.dbp).GetSyncPolicies(ctx, qParams)
	if err != nil {
		log.Error("failed to list sync policies", "error", err)
		return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list sync policies"))
	}
	out := &api.SyncpolicyListOK{}
	for _, row := range pols {
		apiItem, err := qToApiSyncPolicyRow(row)
		if err != nil {
			log.Error("failed to map sync policy row", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to map sync policy row"))
		}
		out.Policies = append(out.Policies, apiItem)
	}
	return out, nil
}

func (h *Handler) SyncpolicyUpdate(ctx context.Context, req *api.SyncPolicy, params api.SyncpolicyUpdateParams) (*api.SyncPolicy, error) {
	log := h.log.With("handler", "SyncpolicyUpdate")
	policyUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("invalid policy uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("invalid sync policy uuid"))
	}
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.SyncPolicy, error) {
		existing, err := query.New(tx).GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
			OrderBy:        nil,
			OrderDirection: "asc",
			Offset:         0,
			Limit:          1,
			Service:        "",
			UUID:           pgtype.UUID{Bytes: uToBytes(policyUUID), Valid: true},
			UserUUID:       "",
			SyncAll:        -1,
		})
		if err != nil {
			log.Error("failed to get sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get sync policy"))
		}
		if len(existing) == 0 {
			return nil, ErrWithCode(http.StatusNotFound, E("sync policy not found"))
		}
		var settingsData []byte
		if req.Settings.IsSet() {
			j, err := json.Marshal(req.Settings.Value)
			if err != nil {
				log.Error("failed to marshal updated sync policy settings", "error", err)
				return nil, ErrWithCode(http.StatusInternalServerError, E("failed to marshal updated sync policy settings"))
			}
			settingsData = j
		} else {
			settingsData = existing[0].Settings
		}
		uParams := query.UpdateSyncPolicyParams{
			UserUUID:    req.UserUUID.Or(""),
			Service:     req.Service,
			Blocklist:   req.Blocklist,
			ExcludeList: req.ExcludeList,
			SyncAll:     req.SyncAll.Or(false),
			Settings:    settingsData,
			UUID:        pgtype.UUID{Bytes: uToBytes(policyUUID), Valid: true},
		}
		err = query.New(tx).UpdateSyncPolicy(ctx, uParams)
		if err != nil {
			log.Error("failed to update sync policy", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update sync policy"))
		}
		outP, err := query.New(tx).GetSyncPolicies(ctx, query.GetSyncPoliciesParams{
			OrderBy:        nil,
			OrderDirection: "asc",
			Offset:         0,
			Limit:          1,
			Service:        "",
			UUID:           pgtype.UUID{Bytes: uToBytes(policyUUID), Valid: true},
			UserUUID:       "",
			SyncAll:        -1,
		})
		if err != nil {
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

func qToApiSyncPolicyRow(row query.GetSyncPoliciesRow) (api.SyncPolicy, error) {
	var sp query.SyncPolicy
	sp.UUID = row.UUID
	sp.UserUUID = row.UserUUID
	sp.Service = row.Service
	sp.Blocklist = row.Blocklist
	sp.ExcludeList = row.ExcludeList
	sp.SyncAll = row.SyncAll
	sp.Settings = row.Settings
	sp.CreatedAt = row.CreatedAt
	sp.UpdatedAt = row.UpdatedAt
	return qToApiSyncPolicy(sp)
}

func qToApiSyncPolicy(dbp query.SyncPolicy) (api.SyncPolicy, error) {
	out := api.SyncPolicy{
		UUID:        api.NewOptString(dbp.UUID.String()),
		Service:     dbp.Service,
		Blocklist:   dbp.Blocklist,
		ExcludeList: dbp.ExcludeList,
		SyncAll:     api.NewOptBool(dbp.SyncAll),
		CreatedAt:   api.NewOptDateTime(dbp.CreatedAt.Time),
		UpdatedAt:   api.NewOptDateTime(dbp.UpdatedAt.Time),
	}
	if dbp.UserUUID != nil {
		out.UserUUID.SetTo(dbp.UserUUID.String())
	}
	if len(dbp.Settings) > 0 {
		var settings map[string]json.RawMessage
		if err := json.Unmarshal(dbp.Settings, &settings); err != nil {
			return out, err
		}
		spSettings := make(api.SyncPolicySettings)
		for k, v := range settings {
			spSettings[k] = jx.Raw(v)
		}
		out.Settings.SetTo(spSettings)
	}
	return out, nil
}

func uToBytes(u uuid.UUID) [16]byte {
	var arr [16]byte
	copy(arr[:], u.Bytes())
	return arr
}

func bytesToUUID(b [16]byte) (uuid.UUID, error) {
	return uuid.FromBytes(b[:])
}
