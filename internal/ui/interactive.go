/*
Package ui manages the terminal user interfaces and logging for the application.
This file provides functions to launch the interactive Text User Interface (TUI) mode,
and helper functions to fetch available models during interactive setup.

Functions:
- RunInteractiveMode: Launches the full-screen TUI and returns the selected setup parameters.
- FetchOllamaModels: Attempts to fetch available Ollama models, failing silently if unreachable.
*/
package ui

import (
	"fmt"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
)

func RunInteractiveMode() (string, []string, string, string, error) {
	rootDir, files, provider, model, err := RunTUI()
	if err != nil {
		return "", nil, "", "", fmt.Errorf("TUI error: %w", err)
	}
	if rootDir == "" || len(files) == 0 {

		return "", nil, "", "", nil
	}
	return rootDir, files, provider, model, nil
}

func FetchOllamaModels(ollamaURL string) []string {
	models, err := llm.GetOllamaModels(ollamaURL)
	if err != nil {

		return []string{}
	}
	return models
}
