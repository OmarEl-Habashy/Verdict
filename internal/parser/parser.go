/*
Package parser provides utilities to parse and handle code extracted from LLM responses.
This file contains the logic needed to safely extract Go code blocks from raw markdown
produced by LLMs, and to write the extracted code to test files in the workspace.

Functions:
- ExtractGoBlock: Extracts a Go code block from a raw LLM response string.
- findClosingFence: Helper to find the index of the closing markdown fence.
- WriteTestFile: Writes the generated test code to a appropriately named _test.go file.
- deriveTestFilePath: Builds the output path for the generated test file.
*/
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ExtractGoBlock(raw string, takeLast bool) (string, error) {

	raw = strings.TrimLeftFunc(raw, func(r rune) bool {
		return r == '\uFEFF' || r == '\u200B' || r == '\u200C' || r == '\u200D'
	})

	raw = strings.ReplaceAll(raw, "```golang", "```go")

	var blocks []string
	remaining := raw
	for {
		start := strings.Index(remaining, "```go")
		if start == -1 {
			break
		}
		rest := remaining[start+5:]

		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}

		end := findClosingFence(rest)
		if end == -1 {
			break
		}
		block := strings.TrimSpace(rest[:end])
		if block != "" {
			blocks = append(blocks, block)
		}

		remaining = rest[end+3:]
	}

	if len(blocks) > 0 {
		if takeLast {
			return blocks[len(blocks)-1], nil
		}
		return blocks[0], nil
	}

	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "package ") {
		return trimmed, nil
	}

	return "", fmt.Errorf("extractGoBlock: no go code block found in LLM response")
}

func findClosingFence(s string) int {
	idx := 0
	for idx < len(s) {
		end := strings.Index(s[idx:], "```")
		if end == -1 {
			return -1
		}
		abs := idx + end

		if abs == 0 || s[abs-1] == '\n' {
			return abs
		}
		idx = abs + 3
	}
	return -1
}

func WriteTestFile(sourcePath, outputDir, code string) (string, error) {
	testPath := deriveTestFilePath(sourcePath, outputDir)

	if err := os.WriteFile(testPath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("writeTestFile: writing %s: %w", testPath, err)
	}
	return testPath, nil
}

func deriveTestFilePath(sourcePath, outputDir string) string {
	base := strings.TrimSuffix(filepath.Base(sourcePath), ".go") + "_test.go"
	if outputDir != "" {
		return filepath.Join(outputDir, base)
	}
	return filepath.Join(filepath.Dir(sourcePath), base)
}
