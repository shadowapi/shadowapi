package mapper

import (
	"github.com/go-faster/jx"
	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// transformRegistry contains all available transform functions.
var transformRegistry = []api.TransformDefinition{
	{
		Type:        "set",
		DisplayName: "Set Value",
		Description: api.NewOptString("Set a static value, ignoring the source field value"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemAny},
		OutputType:  api.TransformDefinitionOutputTypeSame,
		ParamsSchema: api.NewOptTransformDefinitionParamsSchema(mustMakeParamsSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"value": map[string]any{"type": "string", "description": "The static value to set"},
			},
			"required": []string{"value"},
		})),
	},
	{
		Type:        "lowercase",
		DisplayName: "Lowercase",
		Description: api.NewOptString("Convert string to lowercase"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemString},
		OutputType:  api.TransformDefinitionOutputTypeString,
	},
	{
		Type:        "uppercase",
		DisplayName: "Uppercase",
		Description: api.NewOptString("Convert string to uppercase"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemString},
		OutputType:  api.TransformDefinitionOutputTypeString,
	},
	{
		Type:        "trim",
		DisplayName: "Trim Whitespace",
		Description: api.NewOptString("Remove leading and trailing whitespace from string"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemString},
		OutputType:  api.TransformDefinitionOutputTypeString,
	},
	{
		Type:        "to_integer",
		DisplayName: "To Integer",
		Description: api.NewOptString("Parse string as integer number"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemString},
		OutputType:  api.TransformDefinitionOutputTypeInteger,
	},
	{
		Type:        "to_boolean",
		DisplayName: "To Boolean",
		Description: api.NewOptString("Parse string as boolean (accepts: true/false, 1/0, yes/no, on/off)"),
		InputTypes: []api.TransformDefinitionInputTypesItem{
			api.TransformDefinitionInputTypesItemString,
			api.TransformDefinitionInputTypesItemInteger,
		},
		OutputType: api.TransformDefinitionOutputTypeBoolean,
	},
	{
		Type:        "date_parse",
		DisplayName: "Parse Date",
		Description: api.NewOptString("Parse date string using specified format (Go time format)"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemString},
		OutputType:  api.TransformDefinitionOutputTypeDatetime,
		ParamsSchema: api.NewOptTransformDefinitionParamsSchema(mustMakeParamsSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"format": map[string]any{
					"type":        "string",
					"description": "Go time format string (e.g., 2006-01-02 15:04:05)",
				},
				"timezone": map[string]any{
					"type":        "string",
					"description": "Timezone for parsing (e.g., UTC, America/New_York)",
					"default":     "UTC",
				},
			},
			"required": []string{"format"},
		})),
	},
	{
		Type:        "regex_extract",
		DisplayName: "Regex Extract",
		Description: api.NewOptString("Extract value using regular expression capture group"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemString},
		OutputType:  api.TransformDefinitionOutputTypeString,
		ParamsSchema: api.NewOptTransformDefinitionParamsSchema(mustMakeParamsSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "Regular expression pattern with capture groups",
				},
				"group": map[string]any{
					"type":        "integer",
					"description": "Capture group number to extract (1-based, default: 1)",
					"default":     1,
				},
			},
			"required": []string{"pattern"},
		})),
	},
	{
		Type:        "concat",
		DisplayName: "Concatenate",
		Description: api.NewOptString("Concatenate multiple field values or static strings"),
		InputTypes:  []api.TransformDefinitionInputTypesItem{api.TransformDefinitionInputTypesItemAny},
		OutputType:  api.TransformDefinitionOutputTypeString,
		ParamsSchema: api.NewOptTransformDefinitionParamsSchema(mustMakeParamsSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"parts": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type": map[string]any{
								"type":        "string",
								"enum":        []string{"field", "static"},
								"description": "Part type: 'field' to reference a source field, 'static' for literal text",
							},
							"value": map[string]any{
								"type":        "string",
								"description": "Field name (entity.field) or static text depending on type",
							},
						},
						"required": []string{"type", "value"},
					},
					"description": "Parts to concatenate in order",
				},
				"separator": map[string]any{
					"type":        "string",
					"description": "Separator between parts",
					"default":     "",
				},
			},
			"required": []string{"parts"},
		})),
	},
	{
		Type:        "json_extract",
		DisplayName: "JSON Extract",
		Description: api.NewOptString("Extract value from JSON object using path expression"),
		InputTypes: []api.TransformDefinitionInputTypesItem{
			api.TransformDefinitionInputTypesItemObject,
			api.TransformDefinitionInputTypesItemString,
		},
		OutputType: api.TransformDefinitionOutputTypeSame,
		ParamsSchema: api.NewOptTransformDefinitionParamsSchema(mustMakeParamsSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "JSON path expression (e.g., 'user.name' or 'items[0].id')",
				},
			},
			"required": []string{"path"},
		})),
	},
}

// mustMakeParamsSchema converts a map to TransformDefinitionParamsSchema.
func mustMakeParamsSchema(schema map[string]any) api.TransformDefinitionParamsSchema {
	result := make(api.TransformDefinitionParamsSchema)
	for k, v := range schema {
		e := jx.Encoder{}
		encodeAny(&e, v)
		result[k] = e.Bytes()
	}
	return result
}

// encodeAny encodes any value to jx.Encoder.
func encodeAny(e *jx.Encoder, v any) {
	switch val := v.(type) {
	case string:
		e.Str(val)
	case int:
		e.Int(val)
	case bool:
		e.Bool(val)
	case []string:
		e.ArrStart()
		for _, s := range val {
			e.Str(s)
		}
		e.ArrEnd()
	case []any:
		e.ArrStart()
		for _, item := range val {
			encodeAny(e, item)
		}
		e.ArrEnd()
	case map[string]any:
		e.ObjStart()
		for key, value := range val {
			e.FieldStart(key)
			encodeAny(e, value)
		}
		e.ObjEnd()
	default:
		e.Null()
	}
}

// GetAllTransforms returns all available transform functions.
func GetAllTransforms() []api.TransformDefinition {
	return transformRegistry
}

// FilterTransformsByInputType filters transforms by compatible input type.
func FilterTransformsByInputType(transforms []api.TransformDefinition, inputType api.MapperTransformsListInputType) []api.TransformDefinition {
	var targetType api.TransformDefinitionInputTypesItem
	switch inputType {
	case api.MapperTransformsListInputTypeString:
		targetType = api.TransformDefinitionInputTypesItemString
	case api.MapperTransformsListInputTypeInteger:
		targetType = api.TransformDefinitionInputTypesItemInteger
	case api.MapperTransformsListInputTypeBoolean:
		targetType = api.TransformDefinitionInputTypesItemBoolean
	case api.MapperTransformsListInputTypeDatetime:
		targetType = api.TransformDefinitionInputTypesItemDatetime
	case api.MapperTransformsListInputTypeArray:
		targetType = api.TransformDefinitionInputTypesItemArray
	case api.MapperTransformsListInputTypeObject:
		targetType = api.TransformDefinitionInputTypesItemObject
	default:
		return transforms
	}

	result := make([]api.TransformDefinition, 0)
	for _, t := range transforms {
		for _, it := range t.InputTypes {
			if it == targetType || it == api.TransformDefinitionInputTypesItemAny {
				result = append(result, t)
				break
			}
		}
	}
	return result
}

// GetTransformByType returns a transform definition by its type identifier.
func GetTransformByType(transformType string) (api.TransformDefinition, bool) {
	for _, t := range transformRegistry {
		if t.Type == transformType {
			return t, true
		}
	}
	return api.TransformDefinition{}, false
}
