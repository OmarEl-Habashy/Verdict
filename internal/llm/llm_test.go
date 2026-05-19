package llm

import (
	"testing"
)

func TestIsOllamaRunning(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    bool
	}{
		{
			name:    "unreachable host",
			baseURL: "http://localhost:9999",
			want:    false,
		},
		{
			name:    "invalid URL",
			baseURL: "http://invalid-host-that-does-not-exist:11434",
			want:    false,
		},
		{
			name:    "empty URL",
			baseURL: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOllamaRunning(tt.baseURL)
			if got != tt.want {
				t.Errorf("IsOllamaRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOllamaModels(t *testing.T) {
	tests := []struct {
		name      string
		ollamaURL string
		wantErr   bool
	}{
		{
			name:      "unreachable endpoint",
			ollamaURL: "http://localhost:9999/api/chat",
			wantErr:   true,
		},
		{
			name:      "invalid host",
			ollamaURL: "http://invalid-host-that-does-not-exist:11434/api/chat",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetOllamaModels(tt.ollamaURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOllamaModels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAvailableModels(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    []string
	}{
		{
			name:    "unreachable host returns empty slice",
			baseURL: "http://localhost:9999",
			want:    []string{},
		},
		{
			name:    "invalid host returns empty slice",
			baseURL: "http://invalid-host-that-does-not-exist:11434",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableModels(tt.baseURL)
			if len(got) != len(tt.want) {
				t.Errorf("GetAvailableModels() length = %d, want %d", len(got), len(tt.want))
			}
		})
	}
}
