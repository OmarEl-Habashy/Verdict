package parser

import (
	"testing"
)

func TestParseCSVLine(t *testing.T) {
	cases := []struct {
		name      string
		line      string
		want      []string
		wantErr   bool
	}{
		{"valid csv", "field1, field2, field3", []string{"field1", "field2", "field3"}, false},
		{"empty line", "", nil, true},
		{"no fields", ",", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCSVLine(tc.line)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseCSVLine() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !equalStringSlices(got, tc.want) {
				t.Errorf("ParseCSVLine() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{"valid input", "key1=value1\nkey2=value2", map[string]string{"key1": "value1", "key2": "value2"}, false},
		{"empty input", "", map[string]string{}, false},
		{"invalid format", "invalid_format", nil, true},
		{"empty key", "=value", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseKeyValuePairs(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseKeyValuePairs() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !equalStringMaps(got, tc.want) {
				t.Errorf("ParseKeyValuePairs() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		want    []string
		wantErr bool
	}{
		{"valid path", "user.profile.name", []string{"user", "profile", "name"}, false},
		{"empty path", "", nil, true},
		{"no path components", ".", nil, true},
		{"empty path component", "user..name", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseJSONPath(tc.path)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseJSONPath() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !equalStringSlices(got, tc.want) {
				t.Errorf("ParseJSONPath() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractQuotedString(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid quoted string", "\"hello, world\"", "hello, world", false},
		{"missing opening quote", "hello, world\"", "", true},
		{"missing closing quote", "\"hello, world", "", true},
		{"string too short", "\"", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExtractQuotedString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ExtractQuotedString() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("ExtractQuotedString() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSplitOnDelimiter(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		delimiter string
		want      []string
		wantErr   bool
	}{
		{"valid input", "a,b,c", ",", []string{"a", "b", "c"}, false},
		{"empty input", "", ",", []string{}, false},
		{"empty delimiter", "a,b,c", "", nil, true},
		{"no results", "abc", " ", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SplitOnDelimiter(tc.input, tc.delimiter)
			if (err != nil) != tc.wantErr {
				t.Errorf("SplitOnDelimiter() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !equalStringSlices(got, tc.want) {
				t.Errorf("SplitOnDelimiter() = %v, want %v", got, tc.want)
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

func equalStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, val := range a {
		if bVal, exists := b[key]; !exists || bVal != val {
			return false
		}
	}
	return true
}