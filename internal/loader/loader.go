/*
Package loader is responsible for reading and parsing Go source code files.
This file extracts necessary context from a Go file, including package declarations
and exported function names, using string manipulation rather than AST to remain lightweight.

Functions:
- LoadFile: Reads a .go source file, truncates if too large, and returns a populated FileContext.
- ExtractPackageName: Finds and returns the package name declared in the source code.
- ExtractFuncNames: Parses the source code to find all exported function names.
*/
package loader

import (
	"fmt"
	"os"
	"strings"
)

type FileContext struct {
	FilePath    string
	PackageName string
	RawSource   string
	FuncNames   []string
}

func LoadFile(path string) (FileContext, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return FileContext{}, fmt.Errorf("loadFile: reading %s: %w", path, err)
	}

	src := string(raw)
	if len(src) == 0 {
		return FileContext{}, fmt.Errorf("loadFile: file is empty: %s", path)
	}

	src = strings.ReplaceAll(src, "\r\n", "\n")

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

func ExtractFuncNames(src string) []string {
	var names []string
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "func ") {
			continue
		}

		rest := trimmed[len("func "):]

		if strings.HasPrefix(rest, "(") {
			continue
		}

		paren := strings.Index(rest, "(")
		if paren == -1 {
			continue
		}
		name := rest[:paren]

		if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
			names = append(names, name)
		}
	}
	return names
}
