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

func TestNewOpenAIProvider(t *testing.T) {
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
			apiKey:          "sk-test123",
			model:           "gpt-4o",
			timeout:         60 * time.Second,
			expectedModel:   "gpt-4o",
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "with defaults",
			apiKey:          "",
			model:           "",
			timeout:         0,
			expectedModel:   "gpt-4",
			expectedTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewOpenAIProvider(tt.apiKey, tt.model, tt.timeout)

			if provider.model != tt.expectedModel {
				t.Errorf("model = %v, want %v", provider.model, tt.expectedModel)
			}
			if provider.timeout != tt.expectedTimeout {
				t.Errorf("timeout = %v, want %v", provider.timeout, tt.expectedTimeout)
			}
		})
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	provider := NewOpenAIProvider("sk-test", "gpt-4", 30*time.Second)
	if provider.Name() != "openai" {
		t.Errorf("Name() = %v, want openai", provider.Name())
	}
}

func TestOpenAIProvider_Validate(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid key",
			apiKey:  "sk-test123",
			wantErr: false,
		},
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "invalid key format",
			apiKey:  "invalid-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewOpenAIProvider(tt.apiKey, "gpt-4", 30*time.Second)
			err := provider.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenAIProvider_GenerateCommitMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		resp := openAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4",
		}
		resp.Choices = []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "feat: add new feature",
				},
				FinishReason: "stop",
			},
		}
		resp.Usage.TotalTokens = 150

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override the API URL to use mock server
	originalURL := openAIAPIURL
	openAIAPIURL = server.URL
	defer func() { openAIAPIURL = originalURL }()

	provider := NewOpenAIProvider("sk-test", "gpt-4", 30*time.Second)

	ctx := context.Background()
	options := &GenerateOptions{
		MaxTokens:   100,
		Temperature: 0.7,
	}

	message, err := provider.GenerateCommitMessage(ctx, "test diff", options)
	if err != nil {
		t.Fatalf("GenerateCommitMessage() error = %v", err)
	}

	expected := "feat: add new feature"
	if message != expected {
		t.Errorf("GenerateCommitMessage() = %v, want %v", message, expected)
	}
}

func TestOpenAIProvider_GenerateCommitMessage_ValidationError(t *testing.T) {
	provider := NewOpenAIProvider("", "gpt-4", 30*time.Second) // no API key
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

func TestOpenAIProvider_GenerateCommitMessage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		errResp := openAIErrorResponse{}
		errResp.Error.Message = "Internal server error"
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	originalURL := openAIAPIURL
	openAIAPIURL = server.URL
	defer func() { openAIAPIURL = originalURL }()

	provider := NewOpenAIProvider("sk-test", "gpt-4", 30*time.Second)

	ctx := context.Background()
	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}
}

func TestOpenAIProvider_GenerateCommitMessage_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalURL := openAIAPIURL
	openAIAPIURL = server.URL
	defer func() { openAIAPIURL = originalURL }()

	provider := NewOpenAIProvider("sk-test", "gpt-4", 30*time.Second)

	ctx := context.Background()
	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected error for empty response, got nil")
	}
}

func TestOpenAIProvider_GenerateCommitMessage_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		resp := openAIResponse{}
		resp.Choices = []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Content: "feat: test",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	originalURL := openAIAPIURL
	openAIAPIURL = server.URL
	defer func() { openAIAPIURL = originalURL }()

	provider := NewOpenAIProvider("sk-test", "gpt-4", 50*time.Millisecond)
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
