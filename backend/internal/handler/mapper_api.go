package handler

import (
	"context"

	"github.com/shadowapi/shadowapi/backend/internal/mapper"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// MapperSourceFieldsList returns all available source fields for mapping.
func (h *Handler) MapperSourceFieldsList(ctx context.Context, params api.MapperSourceFieldsListParams) (api.MapperSourceFieldsListRes, error) {
	fields := mapper.GetAllSourceFields()

	// Apply datasource type filter first (most restrictive)
	if params.DatasourceType.IsSet() {
		fields = mapper.FilterByDatasourceType(fields, string(params.DatasourceType.Value))
	}

	// Apply entity filter if provided
	if params.Entity.IsSet() {
		fields = mapper.FilterByEntity(fields, params.Entity.Value)
	}

	// Apply type filter if provided
	if params.Type.IsSet() {
		fields = mapper.FilterByType(fields, params.Type.Value)
	}

	return &api.MapperSourceFieldsListOK{Fields: fields}, nil
}

// MapperTransformsList returns all available transform functions.
func (h *Handler) MapperTransformsList(ctx context.Context, params api.MapperTransformsListParams) (api.MapperTransformsListRes, error) {
	transforms := mapper.GetAllTransforms()

	// Apply input type filter if provided
	if params.InputType.IsSet() {
		transforms = mapper.FilterTransformsByInputType(transforms, params.InputType.Value)
	}

	return &api.MapperTransformsListOK{Transforms: transforms}, nil
}

// MapperValidate validates a mapper configuration against source schema and target storage.
func (h *Handler) MapperValidate(ctx context.Context, req *api.MapperValidateReq) (api.MapperValidateRes, error) {
	validator := mapper.NewValidator(h.dbp)
	result := validator.Validate(ctx, req.Config, req.StorageUUID)
	return result, nil
}
