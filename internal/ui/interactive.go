package ui

import (
	"fmt"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
)

// RunInteractiveMode launches the full-screen TUI and returns the selected
// rootDir, file list, provider, and selected model after user confirmation.
// Returns ("", nil, "", "") when the user exits without confirming — caller should exit 0.
func RunInteractiveMode() (string, []string, string, string, error) {
	rootDir, files, provider, model, err := RunTUI()
	if err != nil {
		return "", nil, "", "", fmt.Errorf("TUI error: %w", err)
	}
	if rootDir == "" || len(files) == 0 {
		// User exited or confirmed with no selection — not an error.
		return "", nil, "", "", nil
	}
	return rootDir, files, provider, model, nil
}

// FetchOllamaModels attempts to fetch available Ollama models.
// Returns a list of model names, or an empty slice if Ollama is unreachable.
// This is called during TUI initialization for the local provider.
func FetchOllamaModels(ollamaURL string) []string {
	models, err := llm.GetOllamaModels(ollamaURL)
	if err != nil {
		// Silently fail — models will be empty and TUI will show "no models found"
		return []string{}
	}
	return models
}
