/*
Package ui manages the terminal user interfaces and logging for the application.
This file handles the output logging of test runs to the filesystem, creating
detailed markdown files containing test attempts, outputs, and statuses.

Functions:
- CombineTestOutput: Merges standard output and standard error into a single string.
- LogTestResults: Writes test results to a markdown file in the test directory.
*/
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CombineTestOutput(stdout, stderr string) string {
	output := stdout
	if stderr != "" {
		if output != "" {
			output += "\n\nSTDERR:\n"
		}
		output += stderr
	}
	if output == "" {
		output = "No output recorded"
	}
	return output
}

func LogTestResults(targetFile string, passed bool, testOutput string, attempts int) error {

	status := "success"
	if !passed {
		status = "fail"
	}

	baseName := strings.TrimSuffix(filepath.Base(targetFile), filepath.Ext(targetFile))

	logFileName := fmt.Sprintf("%s_%s_logging.md", status, baseName)
	logPath := filepath.Join(filepath.Dir(targetFile), logFileName)

	content := fmt.Sprintf("# Test Results: %s\n\n", baseName)
	content += fmt.Sprintf("**Status:** %s\n", strings.ToUpper(status))
	content += fmt.Sprintf("**Attempts:** %d\n", attempts)
	content += fmt.Sprintf("**Timestamp:** %s\n\n", time.Now().Format(time.RFC3339))
	content += "## Test Output\n\n```\n"
	content += testOutput
	content += "\n```\n"

	err := os.WriteFile(logPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write test log: %w", err)
	}

	return nil
}
