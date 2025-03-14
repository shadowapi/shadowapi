package handler

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/internal/session"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) DatasourceEmailCreate(ctx context.Context, req *api.DatasourceEmailCreate) (*api.DatasourceEmail, error) {
	log := h.log.With("handler", "DatasourceEmailCreate")
	identity, ok := session.GetIdentity(ctx)
	if !ok {
		log.Error("failed to get identity from context")
		return nil, ErrWithCode(http.StatusUnauthorized, E("not authenticated"))
	}

	userUUID, err := uuid.FromString(identity.ID)
	if err != nil {
		log.Error("failed to parse user uuid", "error", err.Error())
		return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse user uuid"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.DatasourceEmail, error) {
		db := query.New(h.dbp).WithTx(tx)
		cq, err := db.CreateDatasource(ctx, query.CreateDatasourceParams{
			UUID:      uuid.Must(uuid.NewV7()),
			Name:      req.GetName(),
			Type:      "email",
			IsEnabled: req.GetIsEnabled(),
			UserUUID:  &userUUID,
		})
		if err != nil {
			log.Error("failed create base connection object", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed create base connection object"))
		}
		create := query.CreateDatasourceEmailParams{
			UUID:           uuid.Must(uuid.NewV7()),
			DatasourceUUID: &cq.UUID,
			Email:          req.Email,
		}
		if req.Provider.IsSet() {
			create.Provider = req.Provider.Value
		}
		if req.Password.IsSet() {
			create.Password = pgtype.Text{String: req.Password.Value, Valid: true}
		}
		if req.ImapServer.IsSet() {
			create.IMAPServer = pgtype.Text{String: req.ImapServer.Value, Valid: true}
		}
		if req.SMTPServer.IsSet() {
			create.SMTPServer = pgtype.Text{String: req.SMTPServer.Value, Valid: true}
		}
		if req.SMTPTLS.IsSet() {
			create.SMTPTLS = pgtype.Bool{Bool: req.SMTPTLS.Value, Valid: true}
		}
		cqe, err := db.CreateDatasourceEmail(ctx, create)
		if err != nil {
			log.Error("failed create email connection object", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed create email connection object"))
		}
		out := QToDatasource(cq)
		QToDatasourceEmail(&out, cqe)
		return &out, nil
	})
}

func (h *Handler) DatasourceEmailDelete(ctx context.Context, params api.DatasourceEmailDeleteParams) error {
	log := h.log.With("handler", "DatasourceEmailDelete")
	connectionUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse connection uuid", "error", err.Error())
		return ErrWithCode(http.StatusBadRequest, E("failed to parse connection uuid"))
	}

	if err := query.New(h.dbp).DeleteDatasource(ctx, connectionUUID); err != nil {
		log.Error("failed to delete connection", "error", err)
		return ErrWithCode(http.StatusInternalServerError, E("failed to delete connection"))
	}
	return nil
}

func (h *Handler) DatasourceEmailGet(
	ctx context.Context, params api.DatasourceEmailGetParams,
) (*api.DatasourceEmail, error) {
	log := h.log.With("handler", "DatasourceEmailGet")
	connectionUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse connection uuid", "error", err.Error())
		return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse connection uuid"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.DatasourceEmail, error) {
		cq, err := query.New(h.dbp).GetDatasource(ctx, connectionUUID)
		if err != nil {
			log.Error("failed to list connections", "error", err.Error())
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list connections"))
		}
		out := QToDatasource(cq.Datasource)
		QToDatasourceEmail(&out, cq.DatasourceEmail)
		return &out, nil
	})
}

func (h *Handler) DatasourceEmailList(
	ctx context.Context, params api.DatasourceEmailListParams,
) ([]api.DatasourceEmail, error) {
	log := h.log.With("handler", "DatasourceEmailList")
	if _, ok := session.GetIdentity(ctx); !ok {
		log.Error("failed to get identity from context")
		return nil, ErrWithCode(http.StatusUnauthorized, E("not authenticated"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) ([]api.DatasourceEmail, error) {
		lp := query.ListDatasourceParams{}
		if params.Limit.IsSet() {
			lp.LimitRecords = params.Limit.Value
		}
		if params.Offset.IsSet() {
			lp.OffsetRecords = params.Offset.Value
		}
		rows, err := query.New(h.dbp).ListDatasource(ctx, lp)
		if err != nil {
			log.Error("failed to list connections", "error", err.Error())
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list connections"))
		}

		var connections []api.DatasourceEmail
		for _, row := range rows {
			c := QToDatasource(row.Datasource)
			QToDatasourceEmail(&c, row.DatasourceEmail)
			connections = append(connections, c)
		}
		return connections, nil
	})
}

func (h *Handler) DatasourceEmailUpdate(
	ctx context.Context, req *api.DatasourceEmailUpdate, params api.DatasourceEmailUpdateParams,
) (*api.DatasourceEmail, error) {
	log := h.log.With("handler", "DatasourceEmailUpdate", "connectionUUID", params.UUID)

	identity, ok := session.GetIdentity(ctx)
	if !ok {
		log.Error("failed to get identity from context")
		return nil, ErrWithCode(http.StatusUnauthorized, E("not authenticated"))
	}

	userUUID, err := uuid.FromString(identity.ID)
	if err != nil {
		log.Error("failed to parse user uuid", "error", err.Error())
		return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse user uuid"))
	}

	connectionUUID, err := uuid.FromString(params.UUID)
	if err != nil {
		log.Error("failed to parse connection uuid", "error", err)
		return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse connection uuid"))
	}

	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.DatasourceEmail, error) {
		db := query.New(h.dbp).WithTx(tx)
		updateDatasource := query.UpdateDatasourceParams{
			UUID:     connectionUUID,
			UserUUID: &userUUID,
		}
		if req.Name.IsSet() {
			updateDatasource.Name = req.Name.Value
		}
		if req.IsEnabled.IsSet() {
			updateDatasource.IsEnabled = req.IsEnabled.Value
		}
		if req.OAuth2ClientID.IsSet() {
			clientID, err := uuid.FromString(req.OAuth2ClientID.Value)
			if err != nil {
				log.Error("failed to parse oauth2 client id", "error", err)
				return nil, ErrWithCode(http.StatusBadRequest, E("failed to parse oauth2 client id"))
			}
			updateDatasource.OAuth2TokenUUID = &clientID
		}
		if err = db.UpdateDatasource(ctx, updateDatasource); err != nil {
			log.Error("failed to update base connection", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update base connection"))
		}

		updateDatasourceEmail := query.UpdateDatasourceEmailParams{
			DatasourceUUID: &connectionUUID,
		}
		if req.ImapServer.IsSet() {
			updateDatasourceEmail.IMAPServer = pgtype.Text{String: req.ImapServer.Value, Valid: true}
		}
		if req.SMTPServer.IsSet() {
			updateDatasourceEmail.SMTPServer = pgtype.Text{String: req.SMTPServer.Value, Valid: true}
		}
		if req.SMTPTLS.IsSet() {
			updateDatasourceEmail.SMTPTLS = pgtype.Bool{Bool: req.SMTPTLS.Value, Valid: true}
		}
		if err = db.UpdateDatasourceEmail(ctx, updateDatasourceEmail); err != nil {
			log.Error("failed to update email connection", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to update email connection"))
		}
		cq, err := db.GetDatasource(ctx, connectionUUID)
		if err != nil {
			log.Error("failed to get connection", "error", err)
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to get connection"))
		}
		out := QToDatasource(cq.Datasource)
		QToDatasourceEmail(&out, cq.DatasourceEmail)
		return &out, nil
	})
}
