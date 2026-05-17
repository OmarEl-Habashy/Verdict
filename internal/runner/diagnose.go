package runner

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DiagnosticResult describes whether a test failure should trigger healing or abort.
type DiagnosticResult struct {
	ShouldAbort     bool   // true: don't attempt healing, abort immediately
	IsCompilation   bool   // true: test has syntax/structure issues, not worth healing
	IsRuntimeCrash  bool   // true: test crashed (panic, nil map, etc)
	SourceFileError bool   // true: error is in the source file, not the test
	ErrorType       string // from ClassifyError
	AbortReason     string // user-friendly reason for abort
}

// DiagnoseTestFailure performs intelligent triage on a test failure.
// It determines whether healing should be attempted or the run should abort.
// Returns a DiagnosticResult with abort decision and reason.
func DiagnoseTestFailure(result TestResult, sourcePath string) DiagnosticResult {
	diag := DiagnosticResult{
		ErrorType: result.ErrorType,
	}

	// Check error type from combined stderr+stdout.
	combinedErr := result.Stderr + "\n" + result.Stdout
	diag.ErrorType = ClassifyError(combinedErr)

	// Guard 1: Compilation errors should abort immediately (don't waste tokens healing).
	// Compilation errors indicate the test structure itself is broken.
	if diag.ErrorType == "syntax_error" || diag.ErrorType == "compilation" {
		diag.ShouldAbort = true
		diag.IsCompilation = true
		diag.AbortReason = fmt.Sprintf("Compilation error (%s): test has structural issues", diag.ErrorType)
		return diag
	}

	// Guard 2: Runtime crashes should abort immediately (source file issue).
	// Runtime errors like panics indicate the source code itself is broken.
	if diag.ErrorType == "runtime_error" {
		diag.ShouldAbort = true
		diag.IsRuntimeCrash = true
		diag.AbortReason = "Runtime error: source file panic or crash detected"
		return diag
	}

	// Guard 3: Check if error is in source file (not test file).
	// Extract filename from first error line, skip stdlib paths.
	sourceBase := filepath.Base(sourcePath)
	firstFile := extractFirstErrorFile(combinedErr)
	if firstFile != "" && firstFile == sourceBase {
		diag.ShouldAbort = true
		diag.SourceFileError = true
		diag.AbortReason = "Error in source file: cannot heal user code"
		return diag
	}

	// No abort conditions met — this error is worth healing.
	return diag
}

// extractFirstErrorFile extracts the first .go file from error output, skipping stdlib.
// Helper for DiagnoseTestFailure.
func extractFirstErrorFile(errOutput string) string {
	lines := strings.Split(errOutput, "\n")
	for _, line := range lines {
		idx := strings.Index(line, ".go:")
		if idx < 0 {
			continue
		}

		pathPart := strings.TrimSpace(line[:idx+3])
		lower := strings.ToLower(pathPart)

		// Skip stdlib paths.
		if strings.Contains(lower, "/go/src/") ||
			strings.Contains(lower, "\\go\\src\\") ||
			strings.Contains(lower, "program files") ||
			strings.Contains(lower, "/usr/") {
			continue
		}

		return filepath.Base(pathPart)
	}
	return ""
}
