/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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

const (
	claudeAPIURL       = "https://api.anthropic.com/v1/messages"
	defaultClaudeModel = "claude-sonnet-4-5@20250929"
	claudeAPIVersion   = "2023-06-01"
)

// ClaudeProvider implements the Provider interface for Anthropic Claude
type ClaudeProvider struct {
	apiKey  string
	model   string
	timeout time.Duration
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey, model string, timeout time.Duration) *ClaudeProvider {
	if model == "" {
		model = defaultClaudeModel
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ClaudeProvider{
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
	}
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// Validate checks if the provider is properly configured
func (p *ClaudeProvider) Validate() error {
	if p.apiKey == "" {
		return NewProviderError(p.Name(), "API key is required", nil)
	}
	if !strings.HasPrefix(p.apiKey, "sk-ant-") {
		log.Warningln("Claude API key should typically start with 'sk-ant-'")
	}
	return nil
}

// GenerateCommitMessage generates a commit message using Claude
func (p *ClaudeProvider) GenerateCommitMessage(ctx context.Context, diff string, options *GenerateOptions) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}

	prompt := buildPrompt(diff, options)

	reqBody := claudeRequest{
		Model:     p.model,
		MaxTokens: options.MaxTokens,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: options.Temperature,
	}

	// If max tokens not set, use a reasonable default
	if reqBody.MaxTokens == 0 {
		reqBody.MaxTokens = 1024
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to marshal request", err)
	}

	log.Debugln("Claude request payload size:", len(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	client := &http.Client{Timeout: p.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", NewProviderError(p.Name(), "API request failed", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to read response", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp claudeErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return "", NewProviderError(p.Name(), fmt.Sprintf("API error (%d): %s", resp.StatusCode, errResp.Error.Message), nil)
		}
		return "", NewProviderError(p.Name(), fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(body)), nil)
	}

	var apiResp claudeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", NewProviderError(p.Name(), "failed to parse response", err)
	}

	if len(apiResp.Content) == 0 {
		return "", NewProviderError(p.Name(), "no response from API", nil)
	}

	message := strings.TrimSpace(apiResp.Content[0].Text)
	log.Debugln("Claude generated message:", message)
	log.Debugln("Tokens used - input:", apiResp.Usage.InputTokens, "output:", apiResp.Usage.OutputTokens)

	return message, nil
}

// Claude API request/response structures
type claudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []claudeMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type claudeErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
