/*
Package runner executes generated test code and analyzes the results.
This file contains the logic to invoke the `go test` command as a subprocess,
gather its output (including test coverage), and handle cleanup of generated files.

Functions:
- RunTests: Executes the `go test` command with an appropriate timeout and coverage options.
- CleanTestFile: Removes the generated test file after a failed final attempt.
- extractCoverage: Parses the `coverage.out` file to calculate the total coverage percentage.
- parseInt: Safely parses a string to an integer, returning 0 on error.
*/
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

type TestResult struct {
	Passed    bool
	Stdout    string
	Stderr    string
	ExitCode  int
	ErrorType string
	Coverage  float64
}

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

	if strings.Contains(stderrStr, "# ") && strings.Contains(stderrStr, "go vet") {
		stderrStr = "[vet error] " + stderrStr
	}

	if len(stderrStr) > maxStderrLen {
		stderrStr = stderrStr[:maxStderrLen] + "\n... (truncated to 2000 chars)"
	}

	combinedErr := stderrStr + "\n" + stdoutStr
	errType := ClassifyError(combinedErr)

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

func extractCoverage(dir string) float64 {
	coverPath := filepath.Join(dir, "coverage.out")
	data, err := os.ReadFile(coverPath)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return 0
	}

	var totalBlocks, coveredBlocks int

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

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

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
