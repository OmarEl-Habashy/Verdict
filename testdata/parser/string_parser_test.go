package parser

import (
	"testing"
)

func TestParseCSVLine(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		want     []string
		wantErr  bool
	}{
		{"empty line", "", nil, true},
		{"single field", "field", []string{"field"}, false},
		{"multiple fields", "field1, field2", []string{"field1", "field2"}, false},
		{"spaces around fields", "  field1 , field2  ", []string{"field1", "field2"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCSVLine(tc.line)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseCSVLine() error = %v, wantErr %v", err, tc.wantErr)
			} else if !tc.wantErr && !equalSlices(got, tc.want) {
				t.Errorf("ParseCSVLine() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		want     map[string]string
		wantErr  bool
	}{
		{"empty input", "", map[string]string{}, false},
		{"single pair", "key=value", map[string]string{"key": "value"}, false},
		{"multiple pairs", "key1=value1\nkey2=value2", map[string]string{"key1": "value1", "key2": "value2"}, false},
		{"invalid format", "key1=value1,key2", nil, true},
		{"empty key", "=value", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseKeyValuePairs(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseKeyValuePairs() error = %v, wantErr %v", err, tc.wantErr)
			} else if !tc.wantErr && !equalMaps(got, tc.want) {
				t.Errorf("ParseKeyValuePairs() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		want     []string
		wantErr  bool
	}{
		{"empty path", "", nil, true},
		{"valid path", "user.profile.name", []string{"user", "profile", "name"}, false},
		{"path with empty component", "user..name", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseJSONPath(tc.path)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseJSONPath() error = %v, wantErr %v", err, tc.wantErr)
			} else if !tc.wantErr && !equalSlices(got, tc.want) {
				t.Errorf("ParseJSONPath() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExtractQuotedString(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		want     string
		wantErr  bool
	}{
		{"missing opening quote", "hello\"", "", true},
		{"missing closing quote", "\"hello", "", true},
		{"valid quoted string", "\"hello\"", "hello", false},
		{"string with escaped quotes", "\"this is a \\\"test\\\"\"", "this is a \"test\"", false},
		{"string too short", "h", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ExtractQuotedString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ExtractQuotedString() error = %v, wantErr %v", err, tc.wantErr)
			} else if !tc.wantErr && got != tc.want {
				t.Errorf("ExtractQuotedString() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSplitOnDelimiter(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		delimiter string
		want     []string
		wantErr  bool
	}{
		{"empty input", "", ",", []string{}, false},
		{"empty delimiter", "hello,world", "", nil, true},
		{"valid split", "a,b,c", ",", []string{"a", "b", "c"}, false},
		{"split with spaces", " a , b , c ", ",", []string{" a ", " b ", " c "}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SplitOnDelimiter(tc.input, tc.delimiter)
			if (err != nil) != tc.wantErr {
				t.Errorf("SplitOnDelimiter() error = %v, wantErr %v", err, tc.wantErr)
			} else if !tc.wantErr && !equalSlices(got, tc.want) {
				t.Errorf("SplitOnDelimiter() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
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

func equalMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if bValue, ok := b[key]; !ok || value != bValue {
			return false
		}
	}
	return true
}