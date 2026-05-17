package ui

import (
	"fmt"
)

// RunInteractiveMode launches the full-screen TUI and returns the selected
// rootDir and file list after user confirmation.
// Returns ("", nil, nil) when the user exits without confirming — caller should exit 0.
func RunInteractiveMode() (string, []string, string, error) {
	rootDir, files, provider, err := RunTUI()
	if err != nil {
		return "", nil, "", fmt.Errorf("TUI error: %w", err)
	}
	if rootDir == "" || len(files) == 0 {
		// User exited or confirmed with no selection — not an error.
		return "", nil, "", nil
	}
	return rootDir, files, provider, nil
}
