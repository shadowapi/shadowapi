package transforms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Transform represents a transformation configuration.
type Transform struct {
	Type   string
	Params map[string]any
}

// Apply applies a transformation to a value with access to all source data.
// The sourceData is needed for transforms like "concat" that can reference other fields.
func Apply(value any, t Transform, sourceData map[string]any) (any, error) {
	switch t.Type {
	case "set":
		return applySet(t)
	case "lowercase":
		return strings.ToLower(ToString(value)), nil
	case "uppercase":
		return strings.ToUpper(ToString(value)), nil
	case "trim":
		return strings.TrimSpace(ToString(value)), nil
	case "to_integer":
		return strconv.ParseInt(ToString(value), 10, 64)
	case "to_boolean":
		return ParseBool(ToString(value))
	case "date_parse":
		return applyDateParse(value, t)
	case "regex_extract":
		return applyRegexExtract(value, t)
	case "concat":
		return applyConcat(t, sourceData)
	case "json_extract":
		return applyJSONExtract(value, t)
	default:
		return value, nil
	}
}

// applySet returns a static value from params.
func applySet(t Transform) (any, error) {
	if v, ok := GetParamString(t.Params, "value"); ok {
		return v, nil
	}
	return nil, fmt.Errorf("set transform requires 'value' parameter")
}

// applyDateParse parses a date string using the specified format.
func applyDateParse(value any, t Transform) (any, error) {
	format := "2006-01-02 15:04:05"
	if f, ok := GetParamString(t.Params, "format"); ok {
		format = f
	}

	loc := time.UTC
	if tz, ok := GetParamString(t.Params, "timezone"); ok {
		if parsedLoc, err := time.LoadLocation(tz); err == nil {
			loc = parsedLoc
		}
	}

	return time.ParseInLocation(format, ToString(value), loc)
}

// applyRegexExtract extracts a value using regex capture groups.
func applyRegexExtract(value any, t Transform) (any, error) {
	pattern, ok := GetParamString(t.Params, "pattern")
	if !ok {
		return nil, fmt.Errorf("regex_extract requires 'pattern' parameter")
	}

	group := GetParamInt(t.Params, "group", 1)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := re.FindStringSubmatch(ToString(value))
	if len(matches) > group {
		return matches[group], nil
	}
	return "", nil
}

// applyConcat concatenates multiple parts (fields or static values).
func applyConcat(t Transform, sourceData map[string]any) (any, error) {
	parts, ok := GetParamParts(t.Params)
	if !ok {
		return nil, fmt.Errorf("concat requires 'parts' parameter")
	}

	separator := ""
	if s, ok := GetParamString(t.Params, "separator"); ok {
		separator = s
	}

	var resultParts []string
	for _, part := range parts {
		switch part.Type {
		case "field":
			if v, exists := sourceData[part.Value]; exists {
				resultParts = append(resultParts, ToString(v))
			}
		case "static":
			resultParts = append(resultParts, part.Value)
		}
	}

	return strings.Join(resultParts, separator), nil
}

// applyJSONExtract extracts a value from JSON using a path expression.
func applyJSONExtract(value any, t Transform) (any, error) {
	path, ok := GetParamString(t.Params, "path")
	if !ok {
		return nil, fmt.Errorf("json_extract requires 'path' parameter")
	}
	return ExtractJSONPath(value, path)
}

// CollectRequiredFields analyzes mapper mappings to find all required source fields.
// This includes fields referenced in concat transforms.
func CollectRequiredFields(mappings []MappingInfo) map[string]bool {
	required := make(map[string]bool)

	for _, m := range mappings {
		if !m.IsEnabled {
			continue
		}

		// Add the primary source field
		fieldKey := m.SourceEntity + "." + m.SourceField
		required[fieldKey] = true

		// Check for concat transform which may reference other fields
		if m.Transform != nil && m.Transform.Type == "concat" {
			if parts, ok := GetParamParts(m.Transform.Params); ok {
				for _, part := range parts {
					if part.Type == "field" {
						required[part.Value] = true
					}
				}
			}
		}
	}

	return required
}

// MappingInfo is a simplified mapping for field collection.
type MappingInfo struct {
	SourceEntity string
	SourceField  string
	IsEnabled    bool
	Transform    *Transform
}
