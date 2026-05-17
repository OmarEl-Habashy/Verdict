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

// LLMMessage is a single chat message for any LLM backend.
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// --- Ollama wire types ---

// LLMRequest is the request body for the Ollama /api/chat endpoint.
type LLMRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
	Stream   bool         `json:"stream"`
}

// LLMResponse is the response body from Ollama.
type LLMResponse struct {
	Message LLMMessage `json:"message"`
	Error   string     `json:"error,omitempty"`
}

// CallLLM sends a request to an Ollama-compatible /api/chat endpoint.
// timeoutSec is the total HTTP timeout in seconds.
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

// --- OpenAI-compatible wire types ---

// OpenAIRequest is the request body for the OpenAI /v1/chat/completions endpoint.
type OpenAIRequest struct {
	Model    string       `json:"model"`
	Messages []LLMMessage `json:"messages"`
}

// OpenAIChoice wraps a single response choice.
type OpenAIChoice struct {
	Message LLMMessage `json:"message"`
}

// OpenAIResponse is the response body from an OpenAI-compatible endpoint.
type OpenAIResponse struct {
	Choices []OpenAIChoice `json:"choices"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIError holds the error object from an OpenAI-compatible API.
type OpenAIError struct {
	Message string `json:"message"`
}

// CallLLMOpenAI sends a request to an OpenAI-compatible /v1/chat/completions endpoint.
// apiKey is passed as a Bearer token. timeoutSec is the total HTTP timeout.
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
	// OpenRouter requires these headers; free-tier models return 400 without them.
	httpReq.Header.Set("HTTP-Referer", "https://github.com/OmarEl-Habashy/qagent")
	httpReq.Header.Set("X-Title", "qagent")

	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("callLLMOpenAI: connecting to %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Handle rate limiting with a single retry.
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
