package loader

import (
	"fmt"
	"os"
	"strings"
)

// FileContext holds everything extracted from a source .go file.
type FileContext struct {
	FilePath    string
	PackageName string
	RawSource   string
	FuncNames   []string
}

// LoadFile reads a .go source file and returns a populated FileContext.
func LoadFile(path string) (FileContext, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return FileContext{}, fmt.Errorf("loadFile: reading %s: %w", path, err)
	}

	src := string(raw)
	if len(src) == 0 {
		return FileContext{}, fmt.Errorf("loadFile: file is empty: %s", path)
	}

	// Normalize CRLF to LF (Windows source files).
	src = strings.ReplaceAll(src, "\r\n", "\n")

	// Guard against extremely large files that would blow the LLM context.
	const maxBytes = 32000
	if len(src) > maxBytes {
		src = src[:maxBytes] + "\n// ... (source truncated at 32000 bytes)"
	}

	return FileContext{
		FilePath:    path,
		PackageName: ExtractPackageName(src),
		RawSource:   src,
		FuncNames:   ExtractFuncNames(src),
	}, nil
}

// ExtractPackageName returns the package name declared in the source.
// Returns "main" as a fallback if not found.
func ExtractPackageName(src string) string {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "main"
}

// ExtractFuncNames returns the names of all exported functions in the source.
// Uses line-by-line string scanning — no AST, no regex.
func ExtractFuncNames(src string) []string {
	var names []string
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "func ") {
			continue
		}
		// e.g. "func Add(a, b int) int {"
		rest := trimmed[len("func "):]
		// Skip method receivers: "func (r Foo) Method(" starts with "("
		if strings.HasPrefix(rest, "(") {
			continue
		}
		// Extract name up to "("
		paren := strings.Index(rest, "(")
		if paren == -1 {
			continue
		}
		name := rest[:paren]
		// Only exported functions (uppercase first letter).
		if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
			names = append(names, name)
		}
	}
	return names
}
