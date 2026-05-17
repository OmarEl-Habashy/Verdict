package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/OmarEl-Habashy/qagent/internal/loader"
	"github.com/OmarEl-Habashy/qagent/internal/llm"
	"github.com/OmarEl-Habashy/qagent/internal/parser"
	"github.com/OmarEl-Habashy/qagent/internal/runner"
	"github.com/OmarEl-Habashy/qagent/internal/ui"
)

const healCeiling = 5

// Config holds all runtime configuration.
// APIKey must come from QAGENT_API_KEY env var only — never from CLI flags.
type Config struct {
	TargetFile string
	ModelURL   string
	ModelName  string
	MaxHeals   int
	OutputDir  string
	APIKey     string // from QAGENT_API_KEY only
	DryRun     bool
	NoHeal     bool
	Quiet      bool
}

// RunResult is the structured output of a full pipeline run.
type RunResult struct {
	TestFile   string
	Passed     bool
	Attempts   int
	FinalError string
	ElapsedMs  int64
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	if err := checkPrerequisites(); err != nil {
		fmt.Fprintf(os.Stderr, "prerequisite check failed: %s\n", err)
		os.Exit(1)
	}

	start := time.Now()
	result := RunPipeline(cfg)
	elapsed := time.Since(start)
	result.ElapsedMs = elapsed.Milliseconds()

	summary := ui.SummaryResult{
		TestFile:   result.TestFile,
		Passed:     result.Passed,
		Attempts:   result.Attempts,
		FinalError: result.FinalError,
		ElapsedMs:  result.ElapsedMs,
	}

	if cfg.Quiet {
		ui.PrintJSON(summary, elapsed)
	} else {
		ui.PrintSummary(summary, elapsed)
	}

	if !result.Passed {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage:
  qagent --file <path.go> [options]

Options:
  --file        Path to the Go source file (required)
  --model-url   LLM endpoint URL (overrides QAGENT_MODEL_URL)
  --model       Model name (overrides QAGENT_MODEL)
  --max-heals   Max heal attempts 0-5 (default: 2)
  --output-dir  Directory for generated test file
  --dry-run     Print system prompt and exit, do not call LLM
  --no-heal     Disable healing (equivalent to --max-heals 0)
  --quiet       Output JSON only, no colors or spinner

Environment (required):
  QAGENT_API_KEY    API key for the LLM endpoint
  QAGENT_MODEL_URL  LLM endpoint URL
  QAGENT_MODEL      Model name (e.g. openai/gpt-4o-mini)
`)
}

func parseArgs(args []string) (Config, error) {
	cfg := Config{
		ModelURL:  os.Getenv("QAGENT_MODEL_URL"),
		ModelName: os.Getenv("QAGENT_MODEL"),
		MaxHeals:  2,
		APIKey:    os.Getenv("QAGENT_API_KEY"),
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dry-run":
			cfg.DryRun = true
			continue
		case "--no-heal":
			cfg.NoHeal = true
			continue
		case "--quiet":
			cfg.Quiet = true
			continue
		}

		if !strings.HasPrefix(arg, "--") {
			return Config{}, fmt.Errorf("unexpected argument: %q", arg)
		}
		if i+1 >= len(args) {
			return Config{}, fmt.Errorf("flag %s requires a value", arg)
		}
		value := args[i+1]
		i++

		switch arg {
		case "--file":
			abs, err := filepath.Abs(value)
			if err != nil {
				return Config{}, fmt.Errorf("parseArgs: resolving path %q: %w", value, err)
			}
			if _, err := os.Stat(abs); err != nil {
				return Config{}, fmt.Errorf("parseArgs: file not found: %s", abs)
			}
			cfg.TargetFile = abs
		case "--model-url":
			cfg.ModelURL = value
		case "--model":
			cfg.ModelName = value
		case "--max-heals":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("parseArgs: --max-heals must be an integer, got %q", value)
			}
			cfg.MaxHeals = clampMaxHeals(n)
		case "--output-dir":
			cfg.OutputDir = value
		default:
			return Config{}, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	if cfg.TargetFile == "" {
		return Config{}, fmt.Errorf("--file is required")
	}
	if cfg.ModelURL == "" {
		return Config{}, fmt.Errorf("QAGENT_MODEL_URL is not set (use --model-url or set the env var)")
	}
	if cfg.ModelName == "" {
		return Config{}, fmt.Errorf("QAGENT_MODEL is not set (use --model or set the env var)")
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("QAGENT_API_KEY is not set")
	}
	if cfg.NoHeal {
		cfg.MaxHeals = 0
	}

	// Apply qagent.json defaults where CLI values are still at their defaults.
	fileCfg, err := loader.LoadFileConfig("qagent.json")
	if err == nil {
		cliCfg := loader.CLIConfig{
			ModelURL:  cfg.ModelURL,
			ModelName: cfg.ModelName,
			MaxHeals:  cfg.MaxHeals,
			OutputDir: cfg.OutputDir,
			DryRun:    cfg.DryRun,
			NoHeal:    cfg.NoHeal,
			Quiet:     cfg.Quiet,
		}
		merged := loader.MergeConfig(fileCfg, cliCfg)
		cfg.ModelURL = merged.ModelURL
		cfg.ModelName = merged.ModelName
		cfg.MaxHeals = merged.MaxHeals
		cfg.OutputDir = merged.OutputDir
	}

	return cfg, nil
}

func clampMaxHeals(n int) int {
	if n < 0 {
		return 0
	}
	if n > healCeiling {
		return healCeiling
	}
	return n
}

func checkPrerequisites() error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("'go' not found on PATH — install Go: https://go.dev/dl/")
	}
	return nil
}

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
func buildHealPrompt(result runner.TestResult) string {
	errorSection := result.Stderr
	if len(errorSection) > 2000 {
		errorSection = errorSection[:2000] + "\n... (truncated)"
	}
	return fmt.Sprintf(
		"The Go test file you generated failed. Here are the EXACT compiler/runtime errors:\n\n"+
			"--- ERRORS ---\n%s\n--- END ERRORS ---\n\n"+
			"Instructions for fixing:\n"+
			"1. Do NOT change the package name.\n"+
			"2. Do NOT add any new external dependencies.\n"+
			"3. Fix ONLY what the errors above indicate.\n"+
			"4. Output a COMPLETE, corrected ```go code block — not a diff, not a partial snippet.\n"+
			"5. Every test function must start with TestXxx and accept *testing.T.\n\n"+
			"Generate the corrected file now.",
		errorSection,
	)
}

// RunPipeline is the core orchestrator: load → prompt → LLM → parse → test → heal.
func RunPipeline(cfg Config) RunResult {
	if !cfg.Quiet {
		ui.LogStep(1, 4, "Loading source file")
	}

	fileCtx, err := loader.LoadFile(cfg.TargetFile)
	if err != nil {
		ui.LogError("load failed: %s", err)
		return RunResult{Passed: false, FinalError: err.Error()}
	}

	systemPrompt := loader.BuildSystemPrompt(fileCtx)

	if cfg.DryRun {
		fmt.Println(systemPrompt)
		os.Exit(0)
	}

	if !cfg.Quiet {
		ui.LogInfo("File: %s | Package: %s | Funcs: %s",
			filepath.Base(cfg.TargetFile),
			fileCtx.PackageName,
			strings.Join(fileCtx.FuncNames, ", "),
		)
		ui.LogDivider()
	}

	messages := []llm.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Generate the test file now."},
	}

	var lastResult runner.TestResult
	var rawResponse string

	for attempt := 0; attempt <= cfg.MaxHeals; attempt++ {
		if !cfg.Quiet {
			if attempt == 0 {
				ui.LogStep(2, 4, fmt.Sprintf("Calling model: %s", cfg.ModelName))
			} else {
				ui.LogWarn("Healing — attempt %d/%d", attempt+1, cfg.MaxHeals+1)
			}
		}

		spinErr := ui.RunSpinner(
			fmt.Sprintf("Calling %s...", cfg.ModelName),
			func() error {
				var e error
				rawResponse, e = callModel(cfg, messages)
				return e
			},
		)
		if spinErr != nil {
			ui.LogError("LLM call failed: %s", spinErr)
			return RunResult{Passed: false, Attempts: attempt + 1, FinalError: spinErr.Error()}
		}

		takeLast := attempt > 0
		code, parseErr := parser.ExtractGoBlock(rawResponse, takeLast)
		if parseErr != nil {
			ui.LogError("parse error: %s", parseErr)
			return RunResult{Passed: false, Attempts: attempt + 1, FinalError: parseErr.Error()}
		}

		if !cfg.Quiet {
			ui.LogCodeBlock("Generated Test File", code)
			ui.LogStep(3, 4, "Writing test file")
		}

		testPath, writeErr := parser.WriteTestFile(cfg.TargetFile, cfg.OutputDir, code)
		if writeErr != nil {
			ui.LogError("write error: %s", writeErr)
			return RunResult{Passed: false, Attempts: attempt + 1, FinalError: writeErr.Error()}
		}

		if !cfg.Quiet {
			ui.LogStep(4, 4, "Running go test")
		}

		lastResult = runner.RunTests(filepath.Dir(testPath))

		if lastResult.Passed {
			if !cfg.Quiet {
				ui.LogSuccess("Tests passed on attempt %d", attempt+1)
			}
			return RunResult{
				TestFile: testPath,
				Passed:   true,
				Attempts: attempt + 1,
			}
		}

		if !cfg.Quiet {
			ui.LogError("Tests failed (attempt %d/%d)", attempt+1, cfg.MaxHeals+1)
			ui.LogCodeBlock("Compiler Errors", lastResult.Stderr)
		}

		messages = append(messages,
			llm.LLMMessage{Role: "assistant", Content: rawResponse},
			llm.LLMMessage{Role: "user", Content: buildHealPrompt(lastResult)},
		)
	}

	runner.CleanTestFile(cfg.TargetFile, cfg.OutputDir)
	return RunResult{
		Passed:     false,
		Attempts:   cfg.MaxHeals + 1,
		FinalError: lastResult.Stderr,
	}
}
