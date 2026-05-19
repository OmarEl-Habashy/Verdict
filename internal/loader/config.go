/*
Package loader provides functionality to load and merge configurations.
This file defines the structures and logic for reading a qagent.json configuration
file and merging its values with those provided via command-line interfaces.

Functions:
- LoadFileConfig: Reads and parses a qagent.json configuration file.
- MergeConfig: Applies file configuration as defaults, allowing CLI values to take precedence.
*/
package loader

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileConfig struct {
	ModelURL  string `json:"model_url"`
	ModelName string `json:"model_name"`
	MaxHeals  int    `json:"max_heals"`
}

func LoadFileConfig(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, fmt.Errorf("loadFileConfig: reading %s: %w", path, err)
	}

	var cfg FileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FileConfig{}, fmt.Errorf("loadFileConfig: parsing %s: %w", path, err)
	}
	return cfg, nil
}

type CLIConfig struct {
	ModelURL  string
	ModelName string
	MaxHeals  int
	OutputDir string
	DryRun    bool
	NoHeal    bool
	Quiet     bool
}

func MergeConfig(file FileConfig, cli CLIConfig) CLIConfig {
	if cli.ModelURL == "" && file.ModelURL != "" {
		cli.ModelURL = file.ModelURL
	}
	if cli.ModelName == "" && file.ModelName != "" {
		cli.ModelName = file.ModelName
	}

	if file.MaxHeals > 0 && cli.MaxHeals == 2 {
		cli.MaxHeals = file.MaxHeals
	}
	return cli
}
