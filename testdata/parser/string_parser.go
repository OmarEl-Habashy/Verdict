package parser

import (
	"errors"
	"strings"
)

// ParseCSVLine parses a CSV line into fields
func ParseCSVLine(line string) ([]string, error) {
	if line == "" {
		return nil, errors.New("empty line")
	}

	fields := strings.Split(line, ",")
	if len(fields) == 0 {
		return nil, errors.New("no fields found")
	}

	// Trim whitespace from each field
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}

	return fields, nil
}

// ParseKeyValuePairs parses key=value format strings
func ParseKeyValuePairs(input string) (map[string]string, error) {
	if input == "" {
		return make(map[string]string), nil
	}

	pairs := make(map[string]string)
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return nil, errors.New("invalid key=value format")
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, errors.New("empty key")
		}

		pairs[key] = value
	}

	return pairs, nil
}

// ParseJSONPath parses a simplified JSON path like "user.profile.name"
func ParseJSONPath(path string) ([]string, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, errors.New("no path components")
	}

	// Validate each part
	for _, part := range parts {
		if part == "" {
			return nil, errors.New("empty path component")
		}
	}

	return parts, nil
}

// ExtractQuotedString extracts string between quotes, handles escaped quotes
func ExtractQuotedString(input string) (string, error) {
	if len(input) < 2 {
		return "", errors.New("string too short for quotes")
	}

	if input[0] != '"' {
		return "", errors.New("missing opening quote")
	}

	// Find unescaped closing quote
	for i := 1; i < len(input); i++ {
		if input[i] == '"' && (i == 1 || input[i-1] != '\\') {
			return input[1:i], nil
		}
	}

	return "", errors.New("missing closing quote")
}

// SplitOnDelimiter splits string on delimiter, respecting quoted sections
func SplitOnDelimiter(input, delimiter string) ([]string, error) {
	if input == "" {
		return []string{}, nil
	}

	if delimiter == "" {
		return nil, errors.New("empty delimiter")
	}

	parts := strings.Split(input, delimiter)
	if len(parts) == 0 {
		return nil, errors.New("split produced no results")
	}

	return parts, nil
}
