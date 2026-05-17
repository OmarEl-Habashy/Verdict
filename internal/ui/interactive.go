package ui

import (
	"fmt"
)

// RunInteractiveMode displays the welcome banner, launches the TUI file browser,
// and returns the selected rootDir and list of .go files.
// Returns an error if the TUI fails or no files are found.
func RunInteractiveMode() (string, []string, error) {
	fmt.Println()
	colorStep.Println("╔════════════════════════════════════════╗")
	colorStep.Println("║        QAgent — Interactive TUI       ║")
	colorStep.Println("║      Claude Code Vibes Edition        ║")
	colorStep.Println("╚════════════════════════════════════════╝")

	rootDir, allFiles, err := RunTUI()
	if err != nil {
		colorError.Printf("  ✗ TUI failed: %v\n", err)
		return "", nil, err
	}

	if rootDir == "" || len(allFiles) == 0 {
		colorWarn.Println("  ⚠ No files selected")
		return "", nil, fmt.Errorf("no files selected")
	}

	colorSuccess.Printf("\n  ✓ Selected %d files from %s\n\n", len(allFiles), rootDir)

	return rootDir, allFiles, nil
}
