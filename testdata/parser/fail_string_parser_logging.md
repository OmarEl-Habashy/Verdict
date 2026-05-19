# Test Results: string_parser

**Status:** FAIL
**Attempts:** 1
**Timestamp:** 2026-05-19T16:27:11+03:00

## Test Output

```
=== RUN   TestParseCSVLine
=== RUN   TestParseCSVLine/empty_line
=== RUN   TestParseCSVLine/single_field
=== RUN   TestParseCSVLine/multiple_fields
=== RUN   TestParseCSVLine/spaces_around_fields
--- PASS: TestParseCSVLine (0.00s)
    --- PASS: TestParseCSVLine/empty_line (0.00s)
    --- PASS: TestParseCSVLine/single_field (0.00s)
    --- PASS: TestParseCSVLine/multiple_fields (0.00s)
    --- PASS: TestParseCSVLine/spaces_around_fields (0.00s)
=== RUN   TestParseKeyValuePairs
=== RUN   TestParseKeyValuePairs/empty_input
=== RUN   TestParseKeyValuePairs/single_pair
=== RUN   TestParseKeyValuePairs/multiple_pairs
=== RUN   TestParseKeyValuePairs/invalid_format
    string_parser_test.go:50: ParseKeyValuePairs() error = <nil>, wantErr true
=== RUN   TestParseKeyValuePairs/empty_key
--- FAIL: TestParseKeyValuePairs (0.00s)
    --- PASS: TestParseKeyValuePairs/empty_input (0.00s)
    --- PASS: TestParseKeyValuePairs/single_pair (0.00s)
    --- PASS: TestParseKeyValuePairs/multiple_pairs (0.00s)
    --- FAIL: TestParseKeyValuePairs/invalid_format (0.00s)
    --- PASS: TestParseKeyValuePairs/empty_key (0.00s)
=== RUN   TestParseJSONPath
=== RUN   TestParseJSONPath/empty_path
=== RUN   TestParseJSONPath/valid_path
=== RUN   TestParseJSONPath/path_with_empty_component
--- PASS: TestParseJSONPath (0.00s)
    --- PASS: TestParseJSONPath/empty_path (0.00s)
    --- PASS: TestParseJSONPath/valid_path (0.00s)
    --- PASS: TestParseJSONPath/path_with_empty_component (0.00s)
=== RUN   TestExtractQuotedString
=== RUN   TestExtractQuotedString/missing_opening_quote
=== RUN   TestExtractQuotedString/missing_closing_quote
=== RUN   TestExtractQuotedString/valid_quoted_string
=== RUN   TestExtractQuotedString/string_with_escaped_quotes
    string_parser_test.go:102: ExtractQuotedString() got = this is a \"test\", want this is a "test"
=== RUN   TestExtractQuotedString/string_too_short
--- FAIL: TestExtractQuotedString (0.00s)
    --- PASS: TestExtractQuotedString/missing_opening_quote (0.00s)
    --- PASS: TestExtractQuotedString/missing_closing_quote (0.00s)
    --- PASS: TestExtractQuotedString/valid_quoted_string (0.00s)
    --- FAIL: TestExtractQuotedString/string_with_escaped_quotes (0.00s)
    --- PASS: TestExtractQuotedString/string_too_short (0.00s)
=== RUN   TestSplitOnDelimiter
=== RUN   TestSplitOnDelimiter/empty_input
=== RUN   TestSplitOnDelimiter/empty_delimiter
=== RUN   TestSplitOnDelimiter/valid_split
=== RUN   TestSplitOnDelimiter/split_with_spaces
--- PASS: TestSplitOnDelimiter (0.00s)
    --- PASS: TestSplitOnDelimiter/empty_input (0.00s)
    --- PASS: TestSplitOnDelimiter/empty_delimiter (0.00s)
    --- PASS: TestSplitOnDelimiter/valid_split (0.00s)
    --- PASS: TestSplitOnDelimiter/split_with_spaces (0.00s)
FAIL
coverage: 90.0% of statements
FAIL	github.com/OmarEl-Habashy/qagent/testdata/parser	0.739s
FAIL

```
