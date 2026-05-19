/*
Package llm handles the interaction with Large Language Models (LLMs).
It provides structures and functions to construct requests, parse responses,
and handle communication with both local (Ollama) and cloud (OpenAI-compatible) APIs.

Functions:
- CallLLM: Sends a request to an Ollama-compatible /api/chat endpoint.
- CallLLMOpenAI: Sends a request to an OpenAI-compatible /v1/chat/completions endpoint.
- GetAvailableModels: Fetches available models from Ollama safely without returning errors.
- IsOllamaRunning: Checks if an Ollama instance is accessible at the given base URL.
- GetOllamaModels: Fetches the list of available models from a local Ollama instance.
*/
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
	Stream   bool         `json:"stream"`
}

type LLMResponse struct {
	Message LLMMessage `json:"message"`
	Error   string     `json:"error,omitempty"`
}

func CallLLM(url string, req LLMRequest, timeoutSec int) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("callLLM: marshaling request: %w", err)
	}

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("callLLM: connecting to %s: %w — is the model server running?", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("callLLM: server returned %d from %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("callLLM: reading response body: %w", err)
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(data, &llmResp); err != nil {
		return "", fmt.Errorf("callLLM: parsing response: %w", err)
	}
	if llmResp.Error != "" {
		return "", fmt.Errorf("callLLM: model error: %s", llmResp.Error)
	}

	return llmResp.Message.Content, nil
}

type OpenAIRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
}

type OpenAIChoice struct {
	Message LLMMessage `json:"message"`
}

type OpenAIResponse struct {
	Choices []OpenAIChoice `json:"choices"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

type OpenAIError struct {
	Message string `json:"message"`
}

func CallLLMOpenAI(url, apiKey string, req OpenAIRequest, timeoutSec int) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("callLLMOpenAI: marshaling request: %w", err)
	}

	url = strings.TrimSpace(url)
	apiKey = strings.TrimSpace(apiKey)

	httpReq, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, url, bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("callLLMOpenAI: building request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	httpReq.Header.Set("HTTP-Referer", "https://github.com/OmarEl-Habashy/qagent")
	httpReq.Header.Set("X-Title", "qagent")

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("callLLMOpenAI: connecting to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		fmt.Println("Rate limited. Waiting 5s...")
		time.Sleep(5 * time.Second)
		resp.Body.Close()
		resp2, err2 := client.Do(httpReq)
		if err2 != nil {
			return "", fmt.Errorf("callLLMOpenAI: retry after rate limit: %w", err2)
		}
		defer resp2.Body.Close()
		resp = resp2
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("callLLMOpenAI: server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("callLLMOpenAI: reading response: %w", err)
	}

	var oaiResp OpenAIResponse
	if err := json.Unmarshal(data, &oaiResp); err != nil {
		return "", fmt.Errorf("callLLMOpenAI: parsing response: %w", err)
	}
	if oaiResp.Error != nil {
		return "", fmt.Errorf("callLLMOpenAI: API error: %s", oaiResp.Error.Message)
	}
	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("callLLMOpenAI: no choices in response")
	}

	return oaiResp.Choices[0].Message.Content, nil
}

type OllamaModel struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
}

type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

func GetAvailableModels(baseURL string) []string {
	models, err := GetOllamaModels(baseURL)
	if err != nil {
		return []string{}
	}
	return models
}

func IsOllamaRunning(baseURL string) bool {
	tagsURL := strings.TrimSuffix(baseURL, "/") + "/api/tags"
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(tagsURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func GetOllamaModels(ollamaURL string) ([]string, error) {
	tagsURL := strings.TrimSuffix(ollamaURL, "/") + "/api/tags"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(tagsURL)
	if err != nil {
		return nil, fmt.Errorf("getOllamaModels: connecting to %s: %w", tagsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getOllamaModels: server returned %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("getOllamaModels: reading response: %w", err)
	}

	var tagsResp OllamaTagsResponse
	if err := json.Unmarshal(data, &tagsResp); err != nil {
		return nil, fmt.Errorf("getOllamaModels: parsing response: %w", err)
	}

	var modelNames []string
	for _, m := range tagsResp.Models {
		modelNames = append(modelNames, m.Name)
	}
	return modelNames, nil
}
