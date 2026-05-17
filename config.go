package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/OmarEl-Habashy/qagent/internal/loader"
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
	Coverage   bool // if true, collect test coverage with -coverprofile
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
  --coverage    Collect and report test coverage percentage
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
		case "--coverage":
			cfg.Coverage = true
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
