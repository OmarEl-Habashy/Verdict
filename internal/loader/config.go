package loader

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileConfig represents the optional qagent.json config file.
// api_key is intentionally ignored if present — always use QAGENT_API_KEY env var.
type FileConfig struct {
	ModelURL  string `json:"model_url"`
	ModelName string `json:"model_name"`
	MaxHeals  int    `json:"max_heals"`
}

// LoadFileConfig reads and parses a qagent.json file.
// Returns an error if the file does not exist or is malformed.
func LoadFileConfig(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, fmt.Errorf("loadFileConfig: reading %s: %w", path, err)
	}

	// Silently ignore any api_key field by unmarshaling into a raw map first,
	// then re-marshaling without it. Simpler: just unmarshal into FileConfig
	// which has no api_key field — the JSON decoder will ignore unknown keys.
	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, fmt.Errorf("loadFileConfig: parsing %s: %w", path, err)
	}
	return cfg, nil
}

// CLIConfig mirrors the fields in Config that can come from CLI flags.
// Used by MergeConfig to determine which values were explicitly set via CLI.
type CLIConfig struct {
	ModelURL  string
	ModelName string
	MaxHeals  int
	OutputDir string
	DryRun    bool
	NoHeal    bool
	Quiet     bool
}

// MergeConfig applies file config as defaults; CLI values take precedence.
// The caller passes the CLI-resolved config; any zero-value CLI field gets
// the file config value instead.
func MergeConfig(file FileConfig, cli CLIConfig) CLIConfig {
	if cli.ModelURL == "" && file.ModelURL != "" {
		cli.ModelURL = file.ModelURL
	}
	if cli.ModelName == "" && file.ModelName != "" {
		cli.ModelName = file.ModelName
	}
	// MaxHeals: only apply file value if CLI is still at default (0 means explicitly set to 0)
	// We use -1 as sentinel for "not set by user" — but since parseArgs sets default 2,
	// we only override if the user never passed --max-heals (checked by caller).
	if file.MaxHeals > 0 && cli.MaxHeals == 2 {
		cli.MaxHeals = file.MaxHeals
	}
	return cli
}
