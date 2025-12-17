/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package ai

import (
	"testing"
)

func TestProviderError(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		message  string
		err      error
		expected string
	}{
		{
			name:     "Error with wrapped error",
			provider: "openai",
			message:  "API request failed",
			err:      &ProviderError{Provider: "test", Message: "connection timeout"},
			expected: "openai: API request failed: test: connection timeout",
		},
		{
			name:     "Error without wrapped error",
			provider: "claude",
			message:  "Invalid API key",
			err:      nil,
			expected: "claude: Invalid API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProviderError(tt.provider, tt.message, tt.err)
			if err.Error() != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestGenerateOptions(t *testing.T) {
	opts := &GenerateOptions{
		MaxTokens:      1024,
		Temperature:    0.7,
		CommitStandard: "conventional",
		BranchName:     "feat/test-branch",
	}

	if opts.MaxTokens != 1024 {
		t.Errorf("Expected MaxTokens 1024, got %d", opts.MaxTokens)
	}
	if opts.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", opts.Temperature)
	}
	if opts.CommitStandard != "conventional" {
		t.Errorf("Expected CommitStandard 'conventional', got '%s'", opts.CommitStandard)
	}
	if opts.BranchName != "feat/test-branch" {
		t.Errorf("Expected BranchName 'feat/test-branch', got '%s'", opts.BranchName)
	}
}
