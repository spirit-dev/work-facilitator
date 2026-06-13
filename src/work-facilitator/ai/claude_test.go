/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClaudeProvider(t *testing.T) {
	tests := []struct {
		name            string
		apiKey          string
		model           string
		timeout         time.Duration
		expectedModel   string
		expectedTimeout time.Duration
	}{
		{
			name:            "with all parameters",
			apiKey:          "sk-ant-test123",
			model:           "claude-3-opus",
			timeout:         60 * time.Second,
			expectedModel:   "claude-3-opus",
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "with defaults",
			apiKey:          "",
			model:           "",
			timeout:         0,
			expectedModel:   "claude-sonnet-4-5@20250929",
			expectedTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewClaudeProvider(tt.apiKey, tt.model, tt.timeout)

			if provider.model != tt.expectedModel {
				t.Errorf("model = %v, want %v", provider.model, tt.expectedModel)
			}
			if provider.timeout != tt.expectedTimeout {
				t.Errorf("timeout = %v, want %v", provider.timeout, tt.expectedTimeout)
			}
		})
	}
}

func TestClaudeProvider_Name(t *testing.T) {
	provider := NewClaudeProvider("sk-ant-test", "claude-3-opus", 30*time.Second)
	if provider.Name() != "claude" {
		t.Errorf("Name() = %v, want claude", provider.Name())
	}
}

func TestClaudeProvider_Validate(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid key",
			apiKey:  "sk-ant-test123",
			wantErr: false,
		},
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "non-standard prefix (warns but valid)",
			apiKey:  "other-key-format",
			wantErr: false, // Claude only warns, doesn't error on prefix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewClaudeProvider(tt.apiKey, "claude-3-opus", 30*time.Second)
			err := provider.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClaudeProvider_GenerateCommitMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Errorf("Expected anthropic-version header")
		}

		resp := claudeResponse{
			ID:   "msg_123",
			Type: "message",
			Role: "assistant",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{
					Type: "text",
					Text: "fix: resolve bug in auth",
				},
			},
			Model:      "claude-3-opus",
			StopReason: "end_turn",
		}
		resp.Usage.InputTokens = 100
		resp.Usage.OutputTokens = 50

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalURL := claudeAPIURL
	claudeAPIURL = server.URL
	defer func() { claudeAPIURL = originalURL }()

	provider := NewClaudeProvider("sk-ant-test", "claude-3-opus", 30*time.Second)

	ctx := context.Background()
	options := &GenerateOptions{
		MaxTokens:   100,
		Temperature: 0.7,
	}

	message, err := provider.GenerateCommitMessage(ctx, "test diff", options)
	if err != nil {
		t.Fatalf("GenerateCommitMessage() error = %v", err)
	}

	expected := "fix: resolve bug in auth"
	if message != expected {
		t.Errorf("GenerateCommitMessage() = %v, want %v", message, expected)
	}
}

func TestClaudeProvider_GenerateCommitMessage_ValidationError(t *testing.T) {
	provider := NewClaudeProvider("", "claude-3-opus", 30*time.Second) // no API key
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

func TestClaudeProvider_GenerateCommitMessage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		errResp := claudeErrorResponse{}
		errResp.Error.Message = "Internal server error"
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	originalURL := claudeAPIURL
	claudeAPIURL = server.URL
	defer func() { claudeAPIURL = originalURL }()

	provider := NewClaudeProvider("sk-ant-test", "claude-3-opus", 30*time.Second)

	ctx := context.Background()
	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}
}

func TestClaudeProvider_GenerateCommitMessage_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalURL := claudeAPIURL
	claudeAPIURL = server.URL
	defer func() { claudeAPIURL = originalURL }()

	provider := NewClaudeProvider("sk-ant-test", "claude-3-opus", 30*time.Second)

	ctx := context.Background()
	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected error for empty response, got nil")
	}
}

func TestClaudeProvider_GenerateCommitMessage_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		resp := claudeResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{
					Type: "text",
					Text: "feat: test",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalURL := claudeAPIURL
	claudeAPIURL = server.URL
	defer func() { claudeAPIURL = originalURL }()

	provider := NewClaudeProvider("sk-ant-test", "claude-3-opus", 50*time.Millisecond)
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
