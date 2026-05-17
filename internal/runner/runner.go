package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxStderrLen = 2000

// TestResult holds the outcome of a go test run.
type TestResult struct {
	Passed    bool
	Stdout    string
	Stderr    string
	ExitCode  int
	ErrorType string  // "missing_import", "syntax_error", "compilation", "runtime_error", "unknown"
	Coverage  float64 // test coverage percentage (0-100), 0 if not collected
}

// RunTests executes `go test -v -count=1 -timeout=30s ./...` in dir.
// If collectCoverage is true, also passes -coverprofile=coverage.out.
// The subprocess is bounded by a 60s context timeout.
// Stderr is capped at 2000 chars before being stored (see Rule 10).
func RunTests(dir string, collectCoverage bool) TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	args := []string{"test", "-v", "-count=1", "-timeout=30s"}
	if collectCoverage {
		args = append(args, "-coverprofile=coverage.out")
	}
	args = append(args, "./...")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return TestResult{Passed: false, Stderr: "go test timed out after 60s"}
	}

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	stderrStr := stderr.String()
	stdoutStr := stdout.String()

	// Surface go vet errors distinctly.
	if strings.Contains(stderrStr, "# ") && strings.Contains(stderrStr, "go vet") {
		stderrStr = "[vet error] " + stderrStr
	}
	// Cap stderr at 2000 chars to prevent LLM context blowup.
	if len(stderrStr) > maxStderrLen {
		stderrStr = stderrStr[:maxStderrLen] + "\n... (truncated to 2000 chars)"
	}

	// Classify error from both stdout and stderr (panics may be in stdout).
	combinedErr := stderrStr + "\n" + stdoutStr
	errType := ClassifyError(combinedErr)

	// Extract coverage if tests passed and coverage was requested.
	coverage := 0.0
	if exitCode == 0 && collectCoverage {
		coverage = extractCoverage(dir)
	}

	return TestResult{
		Passed:    exitCode == 0,
		Stdout:    stdoutStr,
		Stderr:    stderrStr,
		ExitCode:  exitCode,
		ErrorType: errType,
		Coverage:  coverage,
	}
}

// CleanTestFile removes the generated test file after a failed final attempt.
func CleanTestFile(sourcePath, outputDir string) error {
	base := strings.TrimSuffix(filepath.Base(sourcePath), ".go") + "_test.go"
	var testPath string
	if outputDir != "" {
		testPath = filepath.Join(outputDir, base)
	} else {
		testPath = filepath.Join(filepath.Dir(sourcePath), base)
	}

	if err := os.Remove(testPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cleanTestFile: removing %s: %w", testPath, err)
	}
	return nil
}

// extractCoverage parses coverage.out in dir and returns the total coverage percentage.
// Returns 0 if coverage.out doesn't exist or parsing fails.
func extractCoverage(dir string) float64 {
	coverPath := filepath.Join(dir, "coverage.out")
	data, err := os.ReadFile(coverPath)
	if err != nil {
		return 0 // file doesn't exist or can't be read
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return 0 // empty file
	}

	// Skip header line (mode: set)
	var totalBlocks, coveredBlocks int

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: path/file.go:start.col,end.col numStmt count
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// parts[1] is numStmt, parts[2] is count (1 = covered, 0 = not covered)
		numStmt := parseInt(parts[1])
		count := parseInt(parts[2])

		totalBlocks += numStmt
		if count > 0 {
			coveredBlocks += numStmt
		}
	}

	if totalBlocks == 0 {
		return 0
	}

	return (float64(coveredBlocks) / float64(totalBlocks)) * 100
}

// parseInt safely parses a string to int, returning 0 on error.
func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
