package parser

import (
	"errors"
	"testing"
)

func TestParseCSVLine(t *testing.T) {
	cases := []struct {
		line     string
		expected []string
		err      error
	}{
		{"", nil, errors.New("empty line")},
		{"a,b,c", []string{"a", "b", "c"}, nil},
		{" a , b , c ", []string{"a", "b", "c"}, nil},
		{"a,,c", []string{"a", "", "c"}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.line, func(t *testing.T) {
			got, err := ParseCSVLine(tc.line)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("ParseCSVLine(%q) error = %v; want %v", tc.line, err, tc.err)
			}
			if !equalStringSlices(got, tc.expected) {
				t.Errorf("ParseCSVLine(%q) = %v; want %v", tc.line, got, tc.expected)
			}
		})
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	cases := []struct {
		input    string
		expected map[string]string
		err      error
	}{
		{"", map[string]string{}, nil},
		{"key=value", map[string]string{"key": "value"}, nil},
		{"key1=value1\nkey2=value2", map[string]string{"key1": "value1", "key2": "value2"}, nil},
		{"key1=value1\nkey2= ", map[string]string{"key1": "value1", "key2": ""}, nil},
		{"key1=value1\n=invalid", nil, errors.New("invalid key=value format")},
		{"key1=value1\nkey2", nil, errors.New("invalid key=value format")},
		{"key1=value1\n\nkey2=value2", map[string]string{"key1": "value1", "key2": "value2"}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseKeyValuePairs(tc.input)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("ParseKeyValuePairs(%q) error = %v; want %v", tc.input, err, tc.err)
			}
			if !equalStringMap(got, tc.expected) {
				t.Errorf("ParseKeyValuePairs(%q) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	cases := []struct {
		path     string
		expected []string
		err      error
	}{
		{"", nil, errors.New("empty path")},
		{"user.profile.name", []string{"user", "profile", "name"}, nil},
		{"user..name", nil, errors.New("empty path component")},
		{"user.profile.", nil, errors.New("empty path component")},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got, err := ParseJSONPath(tc.path)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("ParseJSONPath(%q) error = %v; want %v", tc.path, err, tc.err)
			}
			if !equalStringSlices(got, tc.expected) {
				t.Errorf("ParseJSONPath(%q) = %v; want %v", tc.path, got, tc.expected)
			}
		})
	}
}

func TestExtractQuotedString(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		err      error
	}{
		{"", "", errors.New("string too short for quotes")},
		{"no quotes", "", errors.New("missing opening quote")},
		{"\"unquoted", "", errors.New("missing closing quote")},
		{"\"valid quoted\"", "valid quoted", nil},
		{"\"escaped \\\"quote\"", "escaped \"quote", nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ExtractQuotedString(tc.input)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("ExtractQuotedString(%q) error = %v; want %v", tc.input, err, tc.err)
			}
			if got != tc.expected {
				t.Errorf("ExtractQuotedString(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestSplitOnDelimiter(t *testing.T) {
	cases := []struct {
		input    string
		delimiter string
		expected []string
		err      error
	}{
		{"", ",", []string{}, nil},
		{"a,b,c", ",", []string{"a", "b", "c"}, nil},
		{"a;;b;;c", ";", []string{"a", "", "b", "", "c"}, nil},
		{"data|more data|even more data", "|", []string{"data", "more data", "even more data"}, nil},
		{"data|more data|", "|", []string{"data", "more data", ""}, nil},
		{"data|more data|", "", nil, errors.New("empty delimiter")},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := SplitOnDelimiter(tc.input, tc.delimiter)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("SplitOnDelimiter(%q, %q) error = %v; want %v", tc.input, tc.delimiter, err, tc.err)
			}
			if !equalStringSlices(got, tc.expected) {
				t.Errorf("SplitOnDelimiter(%q, %q) = %v; want %v", tc.input, tc.delimiter, got, tc.expected)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}