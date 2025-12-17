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
	openAIAPIURL       = "https://api.openai.com/v1/chat/completions"
	defaultOpenAIModel = "gpt-4"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey  string
	model   string
	timeout time.Duration
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string, timeout time.Duration) *OpenAIProvider {
	if model == "" {
		model = defaultOpenAIModel
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Validate checks if the provider is properly configured
func (p *OpenAIProvider) Validate() error {
	if p.apiKey == "" {
		return NewProviderError(p.Name(), "API key is required", nil)
	}
	if !strings.HasPrefix(p.apiKey, "sk-") {
		return NewProviderError(p.Name(), "invalid API key format", nil)
	}
	return nil
}

// GenerateCommitMessage generates a commit message using OpenAI
func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string, options *GenerateOptions) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}

	prompt := buildPrompt(diff, options)

	reqBody := openAIRequest{
		Model: p.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant that generates concise, meaningful git commit messages based on code changes. Follow conventional commit format when specified.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to marshal request", err)
	}

	log.Debugln("OpenAI request payload size:", len(jsonData))

	req, err := http.NewRequestWithContext(ctx, "POST", openAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
		var errResp openAIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return "", NewProviderError(p.Name(), fmt.Sprintf("API error (%d): %s", resp.StatusCode, errResp.Error.Message), nil)
		}
		return "", NewProviderError(p.Name(), fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(body)), nil)
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", NewProviderError(p.Name(), "failed to parse response", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", NewProviderError(p.Name(), "no response from API", nil)
	}

	message := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	log.Debugln("OpenAI generated message:", message)
	log.Debugln("Tokens used:", apiResp.Usage.TotalTokens)

	return message, nil
}

// OpenAI API request/response structures
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// buildPrompt creates the prompt for the AI based on diff and options
func buildPrompt(diff string, options *GenerateOptions) string {
	var prompt strings.Builder

	prompt.WriteString("Generate a concise git commit message for the following changes:\n\n")
	prompt.WriteString("```diff\n")
	prompt.WriteString(diff)
	prompt.WriteString("\n```\n\n")

	if options.CommitStandard != "" {
		prompt.WriteString(fmt.Sprintf("Follow this commit message standard: %s\n", options.CommitStandard))
	}

	if options.BranchName != "" {
		prompt.WriteString(fmt.Sprintf("Current branch: %s\n", options.BranchName))
	}

	if options.AdditionalContext != "" {
		prompt.WriteString(fmt.Sprintf("Additional context: %s\n", options.AdditionalContext))
	}

	prompt.WriteString("\nProvide ONLY the commit message, without any explanation or additional text. ")
	prompt.WriteString("The message should be a single line (or multiple lines if body is needed). ")
	prompt.WriteString("Do not include markdown formatting or code blocks in your response.")

	return prompt.String()
}
