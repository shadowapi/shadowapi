package handler

import (
	"context"
	"fmt"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

func (h *Handler) PipelineEntryCreate(
	ctx context.Context,
	req *api.PipelineEntryCreateReq,
	params api.PipelineEntryCreateParams,
) (*api.PipelineEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *Handler) PipelineEntryDelete(ctx context.Context, params api.PipelineEntryDeleteParams) error {
	return fmt.Errorf("not implemented")
}

func (h *Handler) PipelineEntryGet(ctx context.Context, params api.PipelineEntryGetParams) (*api.PipelineEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *Handler) PipelineEntryList(
	ctx context.Context,
	params api.PipelineEntryListParams,
) ([]api.PipelineEntry, error) {
	return nil, fmt.Errorf("not implemented")
}

func (h *Handler) PipelineEntryUpdate(
	ctx context.Context,
	req *api.PipelineEntryUpdateReq,
	params api.PipelineEntryUpdateParams,
) (*api.PipelineEntry, error) {
	return nil, fmt.Errorf("not implemented")
}
