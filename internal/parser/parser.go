package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractGoBlock extracts a Go code block from a raw LLM response string.
// On heal attempts (takeLast=true), it returns the LAST block found, not the first —
// the healed version is always appended after the original.
// Falls back to accepting the full raw string if it starts with "package ".
// Never uses regex — pure strings package operations only.
func ExtractGoBlock(raw string, takeLast bool) (string, error) {
	// Strip BOM and zero-width unicode chars that some LLMs emit.
	raw = strings.TrimLeftFunc(raw, func(r rune) bool {
		return r == '\uFEFF' || r == '\u200B' || r == '\u200C' || r == '\u200D'
	})

	// Normalize ```golang to ```go.
	raw = strings.ReplaceAll(raw, "```golang", "```go")

	// Find all ```go blocks and collect them.
	var blocks []string
	remaining := raw
	for {
		start := strings.Index(remaining, "```go")
		if start == -1 {
			break
		}
		rest := remaining[start+5:]
		// Skip the newline immediately after the fence marker.
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
		// Find the closing fence — must be on its own line.
		end := findClosingFence(rest)
		if end == -1 {
			break
		}
		block := strings.TrimSpace(rest[:end])
		if block != "" {
			blocks = append(blocks, block)
		}
		// Advance past the closing fence.
		remaining = rest[end+3:]
	}

	if len(blocks) > 0 {
		if takeLast {
			return blocks[len(blocks)-1], nil
		}
		return blocks[0], nil
	}

	// Fallback: accept raw input if it looks like a Go file.
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "package ") {
		return trimmed, nil
	}

	return "", fmt.Errorf("extractGoBlock: no go code block found in LLM response")
}

// findClosingFence finds the index of the first closing ``` that appears on its own line.
func findClosingFence(s string) int {
	idx := 0
	for idx < len(s) {
		end := strings.Index(s[idx:], "```")
		if end == -1 {
			return -1
		}
		abs := idx + end
		// Check that the ``` is at the start of a line.
		if abs == 0 || s[abs-1] == '\n' {
			return abs
		}
		idx = abs + 3
	}
	return -1
}

// WriteTestFile writes generated test code to a _test.go file.
// If outputDir is non-empty, the file is written there; otherwise it is placed
// in the same directory as the source file.
// Returns the absolute path of the written test file.
func WriteTestFile(sourcePath, outputDir, code string) (string, error) {
	testPath := deriveTestFilePath(sourcePath, outputDir)

	if err := os.WriteFile(testPath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("writeTestFile: writing %s: %w", testPath, err)
	}
	return testPath, nil
}

// deriveTestFilePath builds the output path for the generated test file.
func deriveTestFilePath(sourcePath, outputDir string) string {
	base := strings.TrimSuffix(filepath.Base(sourcePath), ".go") + "_test.go"
	if outputDir != "" {
		return filepath.Join(outputDir, base)
	}
	return filepath.Join(filepath.Dir(sourcePath), base)
}
