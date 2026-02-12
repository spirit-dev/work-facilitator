/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// LlamaCPPProvider implements the Provider interface for llama.cpp server
type LlamaCPPProvider struct {
	baseURL string
	apiKey  string
	model   string
	timeout time.Duration
}

// NewLlamaCPPProvider creates a new LlamaCPP provider instance
func NewLlamaCPPProvider(baseURL, apiKey, model string, timeout time.Duration) *LlamaCPPProvider {
	if baseURL == "" {
		baseURL = "http://localhost:8080/v1"
	}
	if model == "" {
		model = "default"
	}
	if timeout == 0 {
		timeout = 90 * time.Second
	}

	return &LlamaCPPProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
	}
}

// Name returns the provider name
func (p *LlamaCPPProvider) Name() string {
	return "llamacpp"
}

// Validate checks if the provider is properly configured
func (p *LlamaCPPProvider) Validate() error {
	if p.baseURL == "" {
		return NewProviderError("llamacpp", "base URL is required", nil)
	}

	// BaseURL should be a valid URL format
	if !strings.HasPrefix(p.baseURL, "http://") && !strings.HasPrefix(p.baseURL, "https://") {
		return NewProviderError("llamacpp", "base URL must start with http:// or https://", nil)
	}

	log.Debugln("LlamaCPP provider configured with base URL:", p.baseURL)
	return nil
}

// GenerateCommitMessage generates a commit message using llama.cpp
func (p *LlamaCPPProvider) GenerateCommitMessage(ctx context.Context, diff string, options *GenerateOptions) (string, error) {
	if options == nil {
		options = &GenerateOptions{}
	}

	// Build the prompt
	prompt := buildPrompt(diff, options)

	// Prepare request
	reqBody := llamaCPPRequest{
		Model: p.model,
		Messages: []llamaCPPMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant that generates concise, meaningful git commit messages following conventional commit format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
	}

	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewProviderError("llamacpp", "failed to marshal request", err)
	}

	// Create HTTP request
	endpoint := fmt.Sprintf("%s/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", NewProviderError("llamacpp", "failed to create request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	// Send request
	client := &http.Client{Timeout: p.timeout}
	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return "", NewProviderError("llamacpp", "connection refused - is the llama.cpp server running?", err)
		}
		return "", NewProviderError("llamacpp", "request failed", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewProviderError("llamacpp", "failed to read response", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", NewProviderError("llamacpp", fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)), nil)
	}

	// Parse response
	var apiResp llamaCPPResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", NewProviderError("llamacpp", "failed to parse response", err)
	}

	// Extract message
	if len(apiResp.Choices) == 0 {
		return "", NewProviderError("llamacpp", "no choices in response", nil)
	}

	message := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	if message == "" {
		return "", NewProviderError("llamacpp", "empty message in response", nil)
	}

	log.Debugln("LlamaCPP generated message:", message)
	return message, nil
}

// llamaCPPRequest represents the request to llama.cpp server
type llamaCPPRequest struct {
	Model       string            `json:"model"`
	Messages    []llamaCPPMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
}

// llamaCPPMessage represents a message in the conversation
type llamaCPPMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// llamaCPPResponse represents the response from llama.cpp server
type llamaCPPResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []llamaCPPChoice `json:"choices"`
	Usage   llamaCPPUsage    `json:"usage,omitempty"`
}

// llamaCPPChoice represents a completion choice
type llamaCPPChoice struct {
	Index        int             `json:"index"`
	Message      llamaCPPMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// llamaCPPUsage represents token usage information
type llamaCPPUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
