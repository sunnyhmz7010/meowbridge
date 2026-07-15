package webhook

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseJSONPath(source any, path string) (string, bool) {
	if path == "" || !strings.HasPrefix(path, "$") {
		return "", false
	}

	rest := strings.TrimPrefix(path, "$")
	if rest == "" {
		return stringifyJSONPathValue(source)
	}

	current := source
	for rest != "" {
		switch {
		case strings.HasPrefix(rest, "."):
			rest = strings.TrimPrefix(rest, ".")
			name, remaining, ok := consumeProperty(rest)
			if !ok {
				return "", false
			}
			object, ok := current.(map[string]any)
			if !ok {
				return "", false
			}
			value, ok := object[name]
			if !ok {
				return "", false
			}
			current = value
			rest = remaining
		case strings.HasPrefix(rest, "["):
			index, remaining, ok := consumeIndex(rest)
			if !ok {
				return "", false
			}
			values, ok := current.([]any)
			if !ok || index < 0 || index >= len(values) {
				return "", false
			}
			current = values[index]
			rest = remaining
		default:
			return "", false
		}
	}

	return stringifyJSONPathValue(current)
}

func consumeProperty(input string) (string, string, bool) {
	if input == "" {
		return "", "", false
	}
	end := len(input)
	for i, r := range input {
		if r == '.' || r == '[' || r == ']' {
			end = i
			break
		}
	}
	if end == 0 {
		return "", "", false
	}
	return input[:end], input[end:], true
}

func consumeIndex(input string) (int, string, bool) {
	end := strings.Index(input, "]")
	if end <= 1 {
		return 0, "", false
	}
	raw := input[1:end]
	index, err := strconv.Atoi(raw)
	if err != nil {
		return 0, "", false
	}
	return index, input[end+1:], true
}

func stringifyJSONPathValue(value any) (string, bool) {
	if value == nil {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return "", false
		}
		return strings.TrimSpace(typed), true
	default:
		text := strings.TrimSpace(fmt.Sprint(typed))
		return text, text != ""
	}
}
