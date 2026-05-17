package main

import (
	"fmt"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
	"github.com/OmarEl-Habashy/qagent/internal/runner"
)

// callModel dispatches to the appropriate LLM backend based on whether an API key is set.
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

// buildHealPrompt constructs the repair prompt injected after a failed test run.
// Customized based on error type for better LLM guidance.
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
		guidance = "This is a dependency/import error. Do NOT change the logic or package name. Only fix the import statements. " +
			"Check what packages are actually used in the test and import them correctly from the standard library or the source file's imports."
	case "syntax_error":
		guidance = "This is a structural Go syntax error. Check your brackets, parentheses, semicolons, and formatting. " +
			"Common issues: mismatched braces, missing commas in struct definitions, incorrect method receivers."
	case "compilation":
		guidance = "This is a compilation error. Fix ONLY what the error message indicates. " +
			"Do not add new external dependencies or change the package name or function signatures."
	case "runtime_error":
		guidance = "This is a runtime error. The syntax is correct but the logic has issues. " +
			"Check nil pointers, type assertions, and ensure all test cases are valid."
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

// buildCoverageHealPrompt constructs a specialized healing prompt for insufficient test coverage.
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
