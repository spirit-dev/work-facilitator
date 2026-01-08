/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package ai

import (
	"context"
	"time"
)

// Provider defines the interface for AI commit message generation providers
type Provider interface {
	// GenerateCommitMessage generates a commit message based on the git diff
	GenerateCommitMessage(ctx context.Context, diff string, options *GenerateOptions) (string, error)

	// Name returns the provider name
	Name() string

	// Validate checks if the provider is properly configured
	Validate() error
}

// GenerateOptions contains options for commit message generation
type GenerateOptions struct {
	// MaxTokens limits the response length
	MaxTokens int

	// Temperature controls creativity (0.0 to 1.0)
	Temperature float64

	// CommitStandard is the commit message standard to follow (e.g., conventional commits)
	CommitStandard string

	// BranchName is the current branch name for context
	BranchName string

	// AdditionalContext provides extra context for the AI
	AdditionalContext string
}

// Config holds AI provider configuration
type Config struct {
	// Provider name (openai, claude, etc.)
	Provider string

	// APIKey for authentication
	APIKey string

	// Model to use (e.g., gpt-4, claude-3-5-sonnet-20241022)
	Model string

	// BaseURL for providers that support custom endpoints (e.g., llamacpp)
	BaseURL string

	// Enabled indicates if AI features are enabled
	Enabled bool

	// MaxTokens default max tokens for responses
	MaxTokens int

	// Temperature default temperature for generation
	Temperature float64

	// Timeout for API requests
	Timeout time.Duration

	// ExcludePatterns are file patterns to exclude from diff sent to AI
	ExcludePatterns []string
}

// Response represents an AI provider response
type Response struct {
	// Message is the generated commit message
	Message string

	// Provider is the name of the provider that generated the message
	Provider string

	// Model is the specific model used
	Model string

	// TokensUsed is the number of tokens consumed
	TokensUsed int
}

// Error types
type ProviderError struct {
	Provider string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Provider + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Provider + ": " + e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new provider error
func NewProviderError(provider, message string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Message:  message,
		Err:      err,
	}
}
