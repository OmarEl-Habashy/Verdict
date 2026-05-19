/*
Package main provides the entry point for the QAgent CLI application.
This file contains the core logic for communicating with Large Language Models
and generating the specialized prompts required to instruct the LLM on creating
or healing Go test files.

Functions:
- callModel: Dispatches requests to the appropriate LLM backend (local or cloud) based on configuration.
- buildHealPrompt: Constructs a targeted repair prompt after a failed test run, tailored to the specific error type.
- buildCoverageHealPrompt: Constructs a specialized healing prompt when test coverage falls below the required threshold.
*/
package main

import (
	"fmt"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
	"github.com/OmarEl-Habashy/qagent/internal/runner"
)

func callModel(cfg Config, msgs []llm.LLMMessage) (string, error) {
	if cfg.APIKey != "" {
		req := llm.OpenAIRequest{
			Model:    cfg.ModelName,
			Messages: msgs,
		}
		return llm.CallLLMOpenAI(cfg.ModelURL, cfg.APIKey, req, 120)
	}
	req := llm.LLMRequest{
		Model:    cfg.ModelName,
		Messages: msgs,
		Stream:   false,
	}
	return llm.CallLLM(cfg.ModelURL, req, 120)
}

func buildHealPrompt(result runner.TestResult) string {
	if result.ErrorType == "low_coverage" {
		return buildCoverageHealPrompt(result.Coverage)
	}

	errorSection := result.Stderr
	if len(errorSection) > 2000 {
		errorSection = errorSection[:2000] + "\n... (truncated)"
	}

	var guidance string
	switch result.ErrorType {
	case "missing_import":
		guidance = "This is an import error: a function or type is undefined, meaning the package isn't imported.\n" +
			"FIX: Add the missing package to the imports block. Standard library packages (fmt, strings, time, etc.) are always allowed.\n" +
			"Example: If you see 'undefined: fmt', add `import \"fmt\"` to your imports.\n" +
			"Example: If you see 'undefined: strings', add `import \"strings\"` to your imports.\n" +
			"Do NOT remove any logic. Do NOT change the package name."
	case "syntax_error":
		guidance = "This is a structural Go syntax error. Check your brackets, parentheses, semicolons, and formatting. " +
			"Common issues: mismatched braces, missing commas in struct definitions, incorrect method receivers."
	case "compilation":
		guidance = "This is a compilation error. Fix ONLY what the error message indicates. " +
			"Do not add new external dependencies or change the package name or function signatures."
	case "runtime_error":
		guidance = "The test compiled successfully but FAILED at runtime. A test case has an INCORRECT EXPECTATION.\n\n" +
			"CRITICAL AUDIT STEPS:\n" +
			"1. Find the failing test case line in the error output\n" +
			"2. Read the source code for that function carefully\n" +
			"3. Trace through what the function ACTUALLY does with that input\n" +
			"4. Compare: does the test expect the same behavior the function implements?\n\n" +
			"Common mistakes:\n" +
			"- Test expects the function to transform/escape/format, but it doesn't\n" +
			"- Test case value was guessed and doesn't match actual behavior\n" +
			"- Test case is testing a code path that doesn't exist in the source\n\n" +
			"FIX: Update the test case so the expected value (want) matches what the source code ACTUALLY returns.\n" +
			"Do NOT change the source code (the function being tested)."
	default:
		guidance = "This is an unknown error type. Fix ONLY what the errors below indicate. " +
			"Do not change the package name or add new external dependencies."
	}

	return fmt.Sprintf(
		"The Go test file you generated failed. Error type: %s.\n\n"+
			"--- ERRORS ---\n%s\n--- END ERRORS ---\n\n"+
			"Smart guidance for this error type:\n%s\n\n"+
			"General instructions:\n"+
			"1. Do NOT change the package name.\n"+
			"2. Do NOT add any new external dependencies.\n"+
			"3. Output a COMPLETE, corrected ```go code block — not a diff, not a partial snippet.\n"+
			"4. Every test function must start with TestXxx and accept *testing.T.\n\n"+
			"Generate the corrected file now.",
		result.ErrorType,
		errorSection,
		guidance,
	)
}

func buildCoverageHealPrompt(coverage float64) string {
	return fmt.Sprintf(
		"The tests compiled and passed but only achieved %.1f%% coverage. Target is 80%%.\n\n"+
			"Add more test cases to cover:\n"+
			"- All branches of if/else and switch statements\n"+
			"- All error return paths\n"+
			"- Edge cases: zero values, empty strings, nil inputs, negative numbers\n\n"+
			"Output a complete corrected ```go block with additional test cases.\n\n"+
			"General instructions:\n"+
			"1. Do NOT change the package name.\n"+
			"2. Do NOT add any new external dependencies.\n"+
			"3. Output a COMPLETE test file — not a diff, not a partial snippet.\n"+
			"4. Every test function must start with TestXxx and accept *testing.T.\n",
		coverage,
	)
}
