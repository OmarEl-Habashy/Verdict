package runner

import (
	"strings"
)

// ClassifyError categorizes the type of error from stderr/stdout combined.
// Used for smart heal prompt generation.
// Returns one of: "missing_import", "syntax_error", "compilation", "runtime_error", "unknown"
func ClassifyError(output string) string {
	lower := strings.ToLower(output)

	// Check for test failures first: "--- FAIL" indicates the test ran but assertions failed (runtime_error)
	// This must come before syntax checks because test output contains "unexpected" but it's a test assertion, not a syntax error.
	if strings.Contains(output, "--- FAIL") {
		return "runtime_error"
	}

	// Check for import errors (most common compilation issue).
	if strings.Contains(lower, "undefined") || strings.Contains(lower, "could not import") ||
		strings.Contains(lower, "no required module") || strings.Contains(lower, "cannot find package") ||
		strings.Contains(lower, "missing package") || strings.Contains(lower, "no go files") ||
		strings.Contains(lower, "imported and not used") {
		return "missing_import"
	}

	// Check for syntax errors (parser errors, not test assertion failures).
	// "syntax error" is from the parser. "expected" from parser errors is different from test "unexpected".
	if strings.Contains(lower, "syntax error") || strings.Contains(lower, "expected {") ||
		strings.Contains(lower, "expected ;") || strings.Contains(lower, "unexpected token") {
		return "syntax_error"
	}

	// Check for general compilation errors.
	if strings.Contains(lower, "cannot") || strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "not defined") || strings.Contains(lower, "redeclared") {
		return "compilation"
	}

	// Check for runtime errors: panics, nil pointer dereferences, etc.
	if strings.Contains(lower, "panic") || strings.Contains(lower, "fatal error") ||
		strings.Contains(lower, "runtime error") || strings.Contains(lower, "assignment to entry in nil map") {
		return "runtime_error"
	}

	// Everything else.
	return "unknown"
}
