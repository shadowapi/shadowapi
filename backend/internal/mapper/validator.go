package mapper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
	"github.com/shadowapi/shadowapi/backend/pkg/query"
)

// Validator validates mapper configurations.
type Validator struct {
	dbp *pgxpool.Pool
}

// NewValidator creates a new mapper validator.
func NewValidator(dbp *pgxpool.Pool) *Validator {
	return &Validator{dbp: dbp}
}

// ValidationResult holds the result of mapper validation.
type ValidationResult struct {
	IsValid  bool
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error for a specific mapping.
type ValidationError struct {
	MappingID string
	Field     string
	Message   string
}

// ValidationWarning represents a validation warning for a specific mapping.
type ValidationWarning struct {
	MappingID string
	Field     string
	Message   string
}

// Validate validates a mapper configuration against source schema and target storage.
func (v *Validator) Validate(ctx context.Context, config api.MapperConfig, storageUUID api.OptUUID) *api.MapperValidateOK {
	result := &api.MapperValidateOK{
		IsValid:  true,
		Errors:   []api.MapperValidateOKErrorsItem{},
		Warnings: []api.MapperValidateOKWarningsItem{},
	}

	sourceFields := BuildSourceFieldIndex()

	// Load target tables if storage UUID provided
	var targetTables map[string]map[string]api.StoragePostgresField
	if storageUUID.IsSet() {
		var err error
		targetTables, err = v.loadTargetTables(ctx, storageUUID.Value)
		if err != nil {
			result.Errors = append(result.Errors, api.MapperValidateOKErrorsItem{
				MappingID: api.NewOptString(""),
				Field:     api.NewOptString("storage_uuid"),
				Message:   api.NewOptString(fmt.Sprintf("Failed to load storage: %v", err)),
			})
			result.IsValid = false
			return result
		}
	}

	for _, mapping := range config.Mappings {
		mappingID := mapping.ID.Or("")

		// Validate source field exists
		fieldKey := string(mapping.SourceEntity) + "." + mapping.SourceField
		sourceField, exists := sourceFields[fieldKey]
		if !exists {
			result.Errors = append(result.Errors, api.MapperValidateOKErrorsItem{
				MappingID: api.NewOptString(mappingID),
				Field:     api.NewOptString("source_field"),
				Message:   api.NewOptString(fmt.Sprintf("Unknown source field: %s", fieldKey)),
			})
			result.IsValid = false
			continue
		}

		// Validate transform compatibility
		if mapping.Transform.IsSet() {
			if err := v.validateTransform(sourceField, mapping.Transform.Value); err != nil {
				result.Errors = append(result.Errors, api.MapperValidateOKErrorsItem{
					MappingID: api.NewOptString(mappingID),
					Field:     api.NewOptString("transform"),
					Message:   api.NewOptString(err.Error()),
				})
				result.IsValid = false
			}
		}

		// Validate target table/field if storage provided
		if targetTables != nil {
			if err := v.validateTarget(mapping, targetTables); err != nil {
				result.Errors = append(result.Errors, api.MapperValidateOKErrorsItem{
					MappingID: api.NewOptString(mappingID),
					Field:     api.NewOptString("target"),
					Message:   api.NewOptString(err.Error()),
				})
				result.IsValid = false
			}
		}
	}

	return result
}

// loadTargetTables loads target table definitions from storage.
func (v *Validator) loadTargetTables(ctx context.Context, storageUUID uuid.UUID) (map[string]map[string]api.StoragePostgresField, error) {
	q := query.New(v.dbp)

	// Convert google/uuid to pgtype.UUID
	pgUUID := pgtype.UUID{Bytes: storageUUID, Valid: true}

	storageRow, err := q.GetStorage(ctx, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("storage not found: %w", err)
	}

	if storageRow.Storage.Type != "postgres" {
		return nil, fmt.Errorf("storage type must be 'postgres', got '%s'", storageRow.Storage.Type)
	}

	var pgSettings api.StoragePostgres
	if err := json.Unmarshal(storageRow.Storage.Settings, &pgSettings); err != nil {
		return nil, fmt.Errorf("failed to parse storage settings: %w", err)
	}

	result := make(map[string]map[string]api.StoragePostgresField)
	for _, table := range pgSettings.Tables {
		fields := make(map[string]api.StoragePostgresField)
		for _, field := range table.Fields {
			fields[field.Name] = field
		}
		result[table.Name] = fields
	}

	return result, nil
}

// validateTransform checks if the transform is compatible with the source field type.
func (v *Validator) validateTransform(sourceField api.SourceFieldDefinition, transform api.MapperTransform) error {
	transformDef, exists := GetTransformByType(string(transform.Type))
	if !exists {
		return fmt.Errorf("unknown transform type: %s", transform.Type)
	}

	// Check if source field type is compatible with transform input types
	sourceType := sourceFieldTypeToTransformType(sourceField.Type)
	compatible := false
	for _, inputType := range transformDef.InputTypes {
		if inputType == api.TransformDefinitionInputTypesItemAny || inputType == sourceType {
			compatible = true
			break
		}
	}

	if !compatible {
		return fmt.Errorf("transform '%s' does not accept input type '%s'", transform.Type, sourceField.Type)
	}

	return nil
}

// validateTarget checks if the target table and field exist in the storage.
func (v *Validator) validateTarget(mapping api.MapperFieldMapping, targetTables map[string]map[string]api.StoragePostgresField) error {
	tableFields, exists := targetTables[mapping.TargetTable]
	if !exists {
		return fmt.Errorf("target table '%s' not found in storage", mapping.TargetTable)
	}

	_, exists = tableFields[mapping.TargetField]
	if !exists {
		return fmt.Errorf("target field '%s' not found in table '%s'", mapping.TargetField, mapping.TargetTable)
	}

	return nil
}

// sourceFieldTypeToTransformType converts source field type to transform input type.
func sourceFieldTypeToTransformType(fieldType api.SourceFieldDefinitionType) api.TransformDefinitionInputTypesItem {
	switch fieldType {
	case api.SourceFieldDefinitionTypeString:
		return api.TransformDefinitionInputTypesItemString
	case api.SourceFieldDefinitionTypeInteger:
		return api.TransformDefinitionInputTypesItemInteger
	case api.SourceFieldDefinitionTypeBoolean:
		return api.TransformDefinitionInputTypesItemBoolean
	case api.SourceFieldDefinitionTypeDatetime:
		return api.TransformDefinitionInputTypesItemDatetime
	case api.SourceFieldDefinitionTypeArray:
		return api.TransformDefinitionInputTypesItemArray
	case api.SourceFieldDefinitionTypeObject:
		return api.TransformDefinitionInputTypesItemObject
	default:
		return api.TransformDefinitionInputTypesItemAny
	}
}
