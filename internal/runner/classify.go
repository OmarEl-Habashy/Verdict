/*
Package runner executes generated test code and analyzes the results.
This file contains logic to classify and categorize the types of errors
encountered during test runs (such as syntax errors, compilation failures,
missing imports, and runtime errors) by analyzing standard output and error.

Functions:
- ClassifyError: Categorizes the type of error from combined stderr and stdout.
*/
package runner

import (
	"strings"
)

func ClassifyError(output string) string {
	lower := strings.ToLower(output)

	if strings.Contains(output, "--- FAIL") {
		return "runtime_error"
	}

	if strings.Contains(lower, "undefined") || strings.Contains(lower, "could not import") ||
		strings.Contains(lower, "no required module") || strings.Contains(lower, "cannot find package") ||
		strings.Contains(lower, "missing package") || strings.Contains(lower, "no go files") ||
		strings.Contains(lower, "imported and not used") {
		return "missing_import"
	}

	if strings.Contains(lower, "syntax error") || strings.Contains(lower, "expected {") ||
		strings.Contains(lower, "expected ;") || strings.Contains(lower, "unexpected token") {
		return "syntax_error"
	}

	if strings.Contains(lower, "cannot") || strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "not defined") || strings.Contains(lower, "redeclared") {
		return "compilation"
	}

	if strings.Contains(lower, "panic") || strings.Contains(lower, "fatal error") ||
		strings.Contains(lower, "runtime error") || strings.Contains(lower, "assignment to entry in nil map") {
		return "runtime_error"
	}

	return "unknown"
}
