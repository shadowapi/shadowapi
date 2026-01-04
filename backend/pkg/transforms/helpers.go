// Package transforms provides shared transformation functions for mapper execution.
package transforms

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ToString converts any value to string.
func ToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		if v == nil {
			return ""
		}
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

// ParseBool parses a string as boolean.
func ParseBool(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse '%s' as boolean", s)
	}
}

// ExtractJSONPath extracts a value from JSON using a simple path expression.
func ExtractJSONPath(value any, path string) (any, error) {
	// Convert to JSON if needed
	var data map[string]any
	switch v := value.(type) {
	case map[string]any:
		data = v
	case string:
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return nil, fmt.Errorf("cannot parse as JSON: %w", err)
		}
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal to JSON: %w", err)
		}
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, fmt.Errorf("cannot parse as JSON object: %w", err)
		}
	}

	// Navigate the path
	parts := strings.Split(path, ".")
	var current any = data
	for _, part := range parts {
		switch c := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = c[part]
			if !ok {
				return nil, nil // Path not found
			}
		default:
			return nil, nil // Cannot navigate further
		}
	}

	return current, nil
}

// GetParamString extracts a string parameter from params map.
func GetParamString(params map[string]any, key string) (string, bool) {
	if params == nil {
		return "", false
	}
	v, ok := params[key]
	if !ok {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	case []byte:
		var s string
		if err := json.Unmarshal(val, &s); err == nil {
			return s, true
		}
	}
	return "", false
}

// GetParamInt extracts an integer parameter from params map.
func GetParamInt(params map[string]any, key string, defaultVal int) int {
	if params == nil {
		return defaultVal
	}
	v, ok := params[key]
	if !ok {
		return defaultVal
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case []byte:
		var i int
		if err := json.Unmarshal(val, &i); err == nil {
			return i
		}
	}
	return defaultVal
}

// GetParamParts extracts concat parts from params map.
func GetParamParts(params map[string]any) ([]ConcatPart, bool) {
	if params == nil {
		return nil, false
	}
	v, ok := params["parts"]
	if !ok {
		return nil, false
	}

	var parts []ConcatPart

	switch val := v.(type) {
	case []any:
		for _, item := range val {
			if m, ok := item.(map[string]any); ok {
				part := ConcatPart{}
				if t, ok := m["type"].(string); ok {
					part.Type = t
				}
				if v, ok := m["value"].(string); ok {
					part.Value = v
				}
				parts = append(parts, part)
			}
		}
		return parts, len(parts) > 0
	case []byte:
		if err := json.Unmarshal(val, &parts); err == nil {
			return parts, len(parts) > 0
		}
	}
	return nil, false
}

// ConcatPart represents a part in a concat transform.
type ConcatPart struct {
	Type  string `json:"type"`  // "field" or "static"
	Value string `json:"value"` // field name or static text
}
