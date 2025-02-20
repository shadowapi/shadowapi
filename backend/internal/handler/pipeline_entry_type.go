package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/shadowapi/shadowapi/backend/internal/db"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

func (h *Handler) PipelineEntryTypeList(ctx context.Context) (*api.PipelineEntryTypeListOK, error) {
	return db.InTx(ctx, h.dbp, func(tx pgx.Tx) (*api.PipelineEntryTypeListOK, error) {
		db := query.New(h.dbp).WithTx(tx)
		datasources, err := db.ListDatasource(ctx, query.ListDatasourceParams{})
		if err != nil {
			return nil, ErrWithCode(http.StatusInternalServerError, E("failed to list datasources"))
		}

		// add datasources to the result
		result := &api.PipelineEntryTypeListOK{}
		for _, datasource := range datasources {
			result.Entries = append(result.Entries, api.PipelineEntryType{
				UUID:     "017590be-d690-11ef-803b-bf93c7b72ac7",
				Category: "datasource",
				FlowType: "input",
				Name:     datasource.Datasource.Name,
			})
		}

		// // TODO: add extractors
		// result.Entries = append(result.Entries, api.PipelineEntryType{
		// 	UUID:     "0b2e500a-d690-11ef-ae92-efbbecdfcefc",
		// 	Category: "extractor",
		// 	FlowType: "default",
		// 	Name:     "Extract Contact Data",
		// })
		//
		// // TODO: add filters
		// result.Entries = append(result.Entries, api.PipelineEntryType{
		// 	UUID:     "15a06cc6-d690-11ef-8cee-87330f9e1b77",
		// 	Category: "filter",
		// 	FlowType: "default",
		// 	Name:     "Filter Contact Data",
		// })
		//
		// // TODO: add mappers
		// result.Entries = append(result.Entries, api.PipelineEntryType{
		// 	UUID:     "1c23ad38-d690-11ef-9564-035090c817ae",
		// 	Category: "mapper",
		// 	FlowType: "default",
		// 	Name:     "Map Contact Data",
		// })

		// TODO: add storages
		result.Entries = append(result.Entries, api.PipelineEntryType{
			UUID:     "225d7a30-d690-11ef-917d-97ec2553ed2b",
			Category: "storage",
			FlowType: "output",
			Name:     "Store Contact Data",
		})

		return result, nil
	})
}
