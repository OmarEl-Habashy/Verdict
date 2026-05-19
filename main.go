/*
Package main provides the entry point for the QAgent CLI application.
This file handles the initialization, configuration loading, and execution of
both the interactive TUI mode and the headless CLI mode. It orchestrates the
entire core pipeline: prompting the LLM, extracting the generated tests,
running them, and entering the healing loop on failure.

Functions:
- loadEnvFile: Searches for and loads the .env file from the executable directory or parent directories.
- main: The application entry point. Parses arguments and handles the run modes.
- RunPipeline: The core orchestrator orchestrating the LLM generation and healing loop.
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/OmarEl-Habashy/qagent/internal/llm"
	"github.com/OmarEl-Habashy/qagent/internal/loader"
	"github.com/OmarEl-Habashy/qagent/internal/parser"
	"github.com/OmarEl-Habashy/qagent/internal/runner"
	"github.com/OmarEl-Habashy/qagent/internal/ui"
)

type RunResult struct {
	TestFile   string
	Passed     bool
	Attempts   int
	FinalError string
	ErrorType  string
	Coverage   float64
	Stdout     string
	Stderr     string
	ElapsedMs  int64
}

func loadEnvFile() {

	ex, err := os.Executable()
	if err != nil {
		_ = godotenv.Load()
		return
	}
	exePath := filepath.Dir(ex)

	current := exePath
	for {
		envPath := filepath.Join(current, ".env")
		if _, err := os.Stat(envPath); err == nil {

			_ = godotenv.Load(envPath)
			return
		}

		parent := filepath.Dir(current)
		if parent == current {

			break
		}
		current = parent
	}

	_ = godotenv.Load()
}

func main() {
	loadEnvFile()

	if len(os.Args) == 1 {

		for {
			_, allFiles, provider, selectedModel, err := ui.RunInteractiveMode()
			if err != nil {
				os.Exit(1)
			}

			if len(allFiles) == 0 {
				os.Exit(0)
			}

			if provider == "local" {
				modelName := selectedModel
				if modelName == "" {
					modelName = "ollama/mistral"
				}
				os.Setenv("QAGENT_MODEL", modelName)
				os.Setenv("QAGENT_MODEL_URL", "http://localhost:11434/api/chat")
				os.Setenv("QAGENT_API_KEY", "")
			} else if provider == "cloud" {
				if os.Getenv("QAGENT_MODEL") == "" {
					os.Setenv("QAGENT_MODEL", "anthropic/claude-3.5-sonnet")
				}
				if os.Getenv("QAGENT_MODEL_URL") == "" {
					os.Setenv("QAGENT_MODEL_URL", "https://openrouter.ai/api/v1/chat/completions")
				}
			}

			successCount := 0
			failureCount := 0
			returnToMenu := false

			for i, file := range allFiles {
				ui.LogStep(i+1, len(allFiles), fmt.Sprintf("Testing %s", filepath.Base(file)))

				cfg := Config{
					TargetFile: file,
					MaxHeals:   2,
					Coverage:   true,
					ModelName:  os.Getenv("QAGENT_MODEL"),
					ModelURL:   os.Getenv("QAGENT_MODEL_URL"),
					APIKey:     os.Getenv("QAGENT_API_KEY"),
				}
				if cfg.ModelName == "" {
					cfg.ModelName = "ollama/mistral"
				}
				if cfg.ModelURL == "" {
					cfg.ModelURL = "http://localhost:11434/api/chat"
				}

				start := time.Now()
				result := RunPipeline(cfg)
				elapsed := time.Since(start)
				result.ElapsedMs = elapsed.Milliseconds()

				testOutput := ui.CombineTestOutput(result.Stdout, result.Stderr)
				_ = ui.LogTestResults(file, result.Passed, testOutput, result.Attempts)

				if result.Passed {
					successCount++
					ui.LogSuccess("Passed: %s", filepath.Base(file))
				} else {
					failureCount++
					ui.LogError("Failed: %s", filepath.Base(file))

					fmt.Println()
					ui.LogWarn("Diagnostic:")
					diagnostic := ui.PostRunDiagnostic(result.ErrorType, result.FinalError)
					fmt.Println(diagnostic)
				}

				choice := ui.PostRunMenu(result.Passed, result.TestFile, len(allFiles), i)

				if choice == ui.ChoiceNext || choice == ui.ChoiceReturnMenu {

					if !result.Passed && result.TestFile != "" {
						os.Remove(result.TestFile)
					}
				}
				switch choice {
				case ui.ChoiceExtendHeal:
					if !result.Passed {
						fmt.Println()
						ui.LogStep(0, 0, "Attempting 2 additional healing loops...")
						extCfg := cfg
						extCfg.MaxHeals = result.Attempts + 2
						extResult := RunPipeline(extCfg)
						extResult.ElapsedMs = time.Since(start).Milliseconds()

						extTestOutput := ui.CombineTestOutput(extResult.Stdout, extResult.Stderr)
						_ = ui.LogTestResults(file, extResult.Passed, extTestOutput, extResult.Attempts)
						if extResult.Passed {
							successCount++
							failureCount--
							ui.LogSuccess("Passed with extended healing: %s", filepath.Base(file))
							result = extResult
						} else {
							ui.LogError("Still failed after extended healing")
							fmt.Println()
							ui.LogWarn("Diagnostic:")
							diagnostic := ui.PostRunDiagnostic(extResult.ErrorType, extResult.FinalError)
							fmt.Println(diagnostic)

							choice = ui.PostRunMenu(extResult.Passed, extResult.TestFile, len(allFiles), i)

							if (choice == ui.ChoiceNext || choice == ui.ChoiceReturnMenu) && !extResult.Passed && extResult.TestFile != "" {
								os.Remove(extResult.TestFile)
							}
							if choice == ui.ChoiceReturnMenu {
								returnToMenu = true
								break
							}
						}
					}

				case ui.ChoiceReturnMenu:
					returnToMenu = true
					break
				}

				LogRun(RunRecord{
					File:      filepath.Base(file),
					Model:     cfg.ModelName,
					Attempts:  result.Attempts,
					Passed:    result.Passed,
					ErrorType: result.ErrorType,
					Ms:        result.ElapsedMs,
					Timestamp: time.Now().Format(time.RFC3339),
				})
			}

			if returnToMenu {
				fmt.Println()
				ui.LogInfo("Returning to main menu...")
				fmt.Println()
				continue
			}

			fmt.Println()
			ui.LogBatchSummary(successCount, failureCount)
			fmt.Println()
			break
		}
		return
	}

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

	testOutput := ui.CombineTestOutput(result.Stdout, result.Stderr)
	_ = ui.LogTestResults(cfg.TargetFile, result.Passed, testOutput, result.Attempts)

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
		Coverage:   result.Coverage,
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

		lastResult = runner.RunTests(filepath.Dir(testPath), cfg.Coverage)

		if lastResult.Passed && lastResult.Coverage > 0 && lastResult.Coverage < 80.0 {
			lastResult.Passed = false
			lastResult.ErrorType = "low_coverage"
		}

		if lastResult.Passed {
			if !cfg.Quiet {
				ui.LogSuccess("Tests passed on attempt %d", attempt+1)
			}
			return RunResult{
				TestFile: testPath,
				Passed:   true,
				Attempts: attempt + 1,
				Coverage: lastResult.Coverage,
				Stdout:   lastResult.Stdout,
				Stderr:   lastResult.Stderr,
			}
		}

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
				Coverage:   lastResult.Coverage,
				Stdout:     lastResult.Stdout,
				Stderr:     lastResult.Stderr,
			}
		}

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
		Coverage:   lastResult.Coverage,
		Stdout:     lastResult.Stdout,
		Stderr:     lastResult.Stderr,
	}
}
