package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
	"github.com/OmarEl-Habashy/qagent/internal/loader"
	"github.com/OmarEl-Habashy/qagent/internal/parser"
	"github.com/OmarEl-Habashy/qagent/internal/runner"
	"github.com/OmarEl-Habashy/qagent/internal/ui"
)

// RunResult is the structured output of a full pipeline run.
type RunResult struct {
	TestFile   string
	Passed     bool
	Attempts   int
	FinalError string
	ErrorType  string // from runner.DiagnoseTestFailure
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

	// Log run for later analysis.
	LogRun(RunRecord{
		File:      filepath.Base(cfg.TargetFile),
		Model:     cfg.ModelName,
		Attempts:  result.Attempts,
		Passed:    result.Passed,
		ErrorType: result.ErrorType,
		Ms:        result.ElapsedMs,
		Timestamp: time.Now().Format(time.RFC3339),
	})

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

		// Diagnose the failure: should we abort or attempt healing?
		diag := runner.DiagnoseTestFailure(lastResult, cfg.TargetFile)

		if diag.ShouldAbort {
			if !cfg.Quiet {
				ui.LogError(diag.AbortReason)
				errOutput := lastResult.Stderr
				if errOutput == "" && lastResult.Stdout != "" {
					errOutput = lastResult.Stdout
				}
				ui.LogCodeBlock("Error Details", errOutput)
			}
			return RunResult{
				Passed:     false,
				Attempts:   attempt + 1,
				FinalError: lastResult.Stderr,
				ErrorType:  diag.ErrorType,
			}
		}

		// Not aborting — attempt healing.
		if !cfg.Quiet {
			ui.LogError("Tests failed (attempt %d/%d)", attempt+1, cfg.MaxHeals+1)
			errOutput := lastResult.Stderr
			if errOutput == "" && lastResult.Stdout != "" {
				errOutput = lastResult.Stdout
			}
			ui.LogCodeBlock("Compiler Errors", errOutput)
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
		ErrorType:  lastResult.ErrorType,
	}
}
