/*
Package main provides the entry point for the QAgent CLI application.
This file implements the telemetry and usage tracking functionality for QAgent,
saving structured records of each test run to a local JSONL file for later analysis
of failure patterns and model performance.

Functions:
- LogRun: Appends a structured run record to the ~/.qagent/runs.jsonl file, failing silently if an error occurs.
*/
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RunRecord struct {
	File      string `json:"file"`
	Model     string `json:"model"`
	Attempts  int    `json:"attempts"`
	Passed    bool   `json:"passed"`
	ErrorType string `json:"error_type,omitempty"`
	Ms        int64  `json:"ms"`
	Timestamp string `json:"timestamp"`
}

func LogRun(rec RunRecord) {
	homeDir, err := os.UserHomeDir()
	if err != nil {

		return
	}

	logPath := filepath.Join(homeDir, ".qagent", "runs.jsonl")
	dir := filepath.Dir(logPath)

	if err := os.MkdirAll(dir, 0755); err != nil {

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
