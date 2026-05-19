/*
Package runner executes generated test code and analyzes the results.
This file provides diagnostic logic to triage test failures intelligently,
determining whether a failed test is worth "healing" (retrying with the LLM)
or if the run should be immediately aborted due to structural or runtime faults.

Functions:
- DiagnoseTestFailure: Performs triage on a test failure to determine if healing should occur.
- extractFirstErrorFile: Extracts the first .go file involved in an error from the output, skipping the standard library.
*/
package runner

import (
	"fmt"
	"path/filepath"
	"strings"
)

type DiagnosticResult struct {
	ShouldAbort     bool
	IsCompilation   bool
	IsRuntimeCrash  bool
	SourceFileError bool
	ErrorType       string
	AbortReason     string
}

func DiagnoseTestFailure(result TestResult, sourcePath string) DiagnosticResult {
	diag := DiagnosticResult{
		ErrorType: result.ErrorType,
	}

	combinedErr := result.Stderr + "\n" + result.Stdout
	diag.ErrorType = ClassifyError(combinedErr)

	if diag.ErrorType == "syntax_error" || diag.ErrorType == "compilation" {
		diag.ShouldAbort = true
		diag.IsCompilation = true
		diag.AbortReason = fmt.Sprintf("Compilation error (%s): test has structural issues", diag.ErrorType)
		return diag
	}

	if diag.ErrorType == "runtime_error" {
		diag.ShouldAbort = true
		diag.IsRuntimeCrash = true
		diag.AbortReason = "Runtime error: source file panic or crash detected"
		return diag
	}

	sourceBase := filepath.Base(sourcePath)
	firstFile := extractFirstErrorFile(combinedErr)
	if firstFile != "" && firstFile == sourceBase {
		diag.ShouldAbort = true
		diag.SourceFileError = true
		diag.AbortReason = "Error in source file: cannot heal user code"
		return diag
	}

	return diag
}

func extractFirstErrorFile(errOutput string) string {
	lines := strings.Split(errOutput, "\n")
	for _, line := range lines {
		idx := strings.Index(line, ".go:")
		if idx < 0 {
			continue
		}

		pathPart := strings.TrimSpace(line[:idx+3])
		lower := strings.ToLower(pathPart)

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
