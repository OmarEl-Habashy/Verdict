package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const maxStderrLen = 2000

// TestResult holds the outcome of a go test run.
type TestResult struct {
	Passed   bool
	Stdout   string
	Stderr   string
	ExitCode int
}

// RunTests executes `go test -v -count=1 -timeout=30s ./...` in dir.
// The subprocess is bounded by a 60s context timeout.
// Stderr is capped at 2000 chars before being stored (see Rule 9).
func RunTests(dir string) TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "-v", "-count=1", "-timeout=30s", "./...")
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
	// Surface go vet errors distinctly.
	if strings.Contains(stderrStr, "# ") && strings.Contains(stderrStr, "go vet") {
		stderrStr = "[vet error] " + stderrStr
	}
	// Cap stderr at 2000 chars to prevent LLM context blowup.
	if len(stderrStr) > maxStderrLen {
		stderrStr = stderrStr[:maxStderrLen] + "\n... (truncated to 2000 chars)"
	}

	return TestResult{
		Passed:   exitCode == 0,
		Stdout:   stdout.String(),
		Stderr:   stderrStr,
		ExitCode: exitCode,
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
