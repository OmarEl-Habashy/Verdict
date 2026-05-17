package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RunRecord is logged to ~/.qagent/runs.jsonl after each run.
// Use for later analysis of failure patterns and model performance.
type RunRecord struct {
	File      string `json:"file"`
	Model     string `json:"model"`
	Attempts  int    `json:"attempts"`
	Passed    bool   `json:"passed"`
	ErrorType string `json:"error_type,omitempty"`
	Ms        int64  `json:"ms"`
	Timestamp string `json:"timestamp"`
}

// LogRun appends a structured run record to ~/.qagent/runs.jsonl.
// Non-fatal: errors are logged but do not affect the main pipeline.
func LogRun(rec RunRecord) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Silently fail — logging is non-critical.
		return
	}

	logPath := filepath.Join(homeDir, ".qagent", "runs.jsonl")
	dir := filepath.Dir(logPath)

	// Create directory if it doesn't exist.
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Silently fail — logging is non-critical.
		return
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	data, _ := json.Marshal(rec)
	f.Write(append(data, '\n'))
}
