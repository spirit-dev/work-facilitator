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

func TestNewLlamaCPPProvider(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		apiKey          string
		model           string
		timeout         time.Duration
		expectedBaseURL string
		expectedModel   string
		expectedTimeout time.Duration
	}{
		{
			name:            "with all parameters",
			baseURL:         "http://localhost:8080/v1",
			apiKey:          "test-key",
			model:           "llama-2-7b",
			timeout:         60 * time.Second,
			expectedBaseURL: "http://localhost:8080/v1",
			expectedModel:   "llama-2-7b",
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "with defaults",
			baseURL:         "",
			apiKey:          "",
			model:           "",
			timeout:         0,
			expectedBaseURL: "http://localhost:8080/v1",
			expectedModel:   "default",
			expectedTimeout: 90 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewLlamaCPPProvider(tt.baseURL, tt.apiKey, tt.model, tt.timeout)

			if provider.baseURL != tt.expectedBaseURL {
				t.Errorf("baseURL = %v, want %v", provider.baseURL, tt.expectedBaseURL)
			}
			if provider.model != tt.expectedModel {
				t.Errorf("model = %v, want %v", provider.model, tt.expectedModel)
			}
			if provider.timeout != tt.expectedTimeout {
				t.Errorf("timeout = %v, want %v", provider.timeout, tt.expectedTimeout)
			}
		})
	}
}

func TestLlamaCPPProvider_Name(t *testing.T) {
	provider := NewLlamaCPPProvider("http://localhost:8080/v1", "", "default", 90*time.Second)
	if provider.Name() != "llamacpp" {
		t.Errorf("Name() = %v, want llamacpp", provider.Name())
	}
}

func TestLlamaCPPProvider_Validate(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			baseURL: "http://localhost:8080/v1",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			baseURL: "https://localhost:8080/v1",
			wantErr: false,
		},
		{
			name:    "empty URL",
			baseURL: "",
			wantErr: true,
		},
		{
			name:    "invalid URL scheme",
			baseURL: "ftp://localhost:8080/v1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewLlamaCPPProvider(tt.baseURL, "", "default", 90*time.Second)
			// If we want to test empty URL, we must force it because constructor sets default
			if tt.baseURL == "" {
				provider.baseURL = ""
			}
			err := provider.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLlamaCPPProvider_GenerateCommitMessage_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected /chat/completions path, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Send mock response
		resp := llamaCPPResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "default",
			Choices: []llamaCPPChoice{
				{
					Index: 0,
					Message: llamaCPPMessage{
						Role:    "assistant",
						Content: "feat: add new feature",
					},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create provider with mock server URL
	// Note: httptest.Server URL includes http:// prefix
	provider := NewLlamaCPPProvider(server.URL, "", "default", 30*time.Second)

	// Test generation
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

func TestLlamaCPPProvider_GenerateCommitMessage_WithAuth(t *testing.T) {
	expectedKey := "test-api-key"

	// Create mock server that checks auth header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + expectedKey
		if authHeader != expectedAuth {
			t.Errorf("Expected Authorization header %s, got %s", expectedAuth, authHeader)
		}

		resp := llamaCPPResponse{
			Choices: []llamaCPPChoice{
				{
					Message: llamaCPPMessage{
						Content: "feat: test",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewLlamaCPPProvider(server.URL, expectedKey, "default", 30*time.Second)
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err != nil {
		t.Fatalf("GenerateCommitMessage() error = %v", err)
	}
}

func TestLlamaCPPProvider_GenerateCommitMessage_NoAuth(t *testing.T) {
	// Create mock server that checks no auth header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			t.Errorf("Expected no Authorization header, got %s", authHeader)
		}

		resp := llamaCPPResponse{
			Choices: []llamaCPPChoice{
				{
					Message: llamaCPPMessage{
						Content: "feat: test",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewLlamaCPPProvider(server.URL, "", "default", 30*time.Second)
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err != nil {
		t.Fatalf("GenerateCommitMessage() error = %v", err)
	}
}

func TestLlamaCPPProvider_GenerateCommitMessage_ServerError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	provider := NewLlamaCPPProvider(server.URL, "", "default", 30*time.Second)
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}
}

func TestLlamaCPPProvider_GenerateCommitMessage_EmptyResponse(t *testing.T) {
	// Create mock server with empty choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := llamaCPPResponse{
			Choices: []llamaCPPChoice{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewLlamaCPPProvider(server.URL, "", "default", 30*time.Second)
	ctx := context.Background()

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected error for empty response, got nil")
	}
}

func TestLlamaCPPProvider_GenerateCommitMessage_Timeout(t *testing.T) {
	// Create mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		resp := llamaCPPResponse{
			Choices: []llamaCPPChoice{
				{
					Message: llamaCPPMessage{
						Content: "feat: test",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Set timeout shorter than delay
	provider := NewLlamaCPPProvider(server.URL, "", "default", 50*time.Millisecond)
	ctx := context.Background()

	// We need to pass a context with timeout that matches the provider timeout
	// or rely on the client timeout. The provider implementation uses client timeout.
	// However, GenerateCommitMessage doesn't enforce timeout on the context passed to it,
	// it relies on the caller to pass a context with timeout OR the http client timeout.
	// In our implementation: client := &http.Client{Timeout: p.timeout}

	_, err := provider.GenerateCommitMessage(ctx, "test diff", &GenerateOptions{})
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
