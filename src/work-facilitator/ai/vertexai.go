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
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	defaultVertexAIModel    = "gemini-2.5-flash"
	defaultVertexAILocation = "us-central1"
	vertexAIScope           = "https://www.googleapis.com/auth/cloud-platform"
	vertexAITokenURL        = "https://oauth2.googleapis.com/token"
)

// VertexAIProvider implements the Provider interface for Google Vertex AI
type VertexAIProvider struct {
	projectID         string
	location          string
	publisher         string // Publisher for the model (e.g., "google", "google-anthropic-claude")
	model             string // Model name (e.g., "gemini-2.5-flash", "claude-sonnet-4.5-20250517")
	tokenSource       oauth2.TokenSource
	timeout           time.Duration

}



// ParseModelString parses a model string into publisher and model components.
// The model string can be in one of two formats:
//   - "publisher/model" - e.g., "google-anthropic-claude/claude-sonnet-4.5-20250517"
//   - "model" - e.g., "gemini-2.5-flash" (defaults to "google" publisher)
//
// If the model string contains multiple "/" characters, only the first one is used
// as the separator between publisher and model. For example:
//   - "custom/model/with/slashes" → publisher: "custom", model: "model/with/slashes"
//
// Returns the publisher and model as separate strings.
func ParseModelString(modelStr string) (publisher, model string) {
	// Check if the model string contains a "/" separator
	if idx := strings.Index(modelStr, "/"); idx != -1 {
		// Split on the first "/" only
		publisher = modelStr[:idx]
		model = modelStr[idx+1:]
	} else {
		// No "/" found, use default publisher
		publisher = "google"
		model = modelStr
	}

	log.Debugln("Parsed model string: input=", modelStr, ", publisher=", publisher, ", model=", model)
	return publisher, model
}

// NewVertexAIProvider creates a new Vertex AI provider
func NewVertexAIProvider(projectID, location, serviceAccountKeyPath, model string, timeout time.Duration) (*VertexAIProvider, error) {
	if model == "" {
		model = defaultVertexAIModel
	}
	if location == "" {
		location = defaultVertexAILocation
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Parse model string to extract publisher and model
	publisher, parsedModel := ParseModelString(model)
	log.Debugln("Vertex AI provider initialized with publisher:", publisher, "model:", parsedModel)

	// Read credentials file
	jsonData, err := os.ReadFile(serviceAccountKeyPath)
	if err != nil {
		return nil, NewProviderError("vertexai", "failed to read credentials file", err)
	}

	// Load Google credentials
	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, jsonData, vertexAIScope)
	if err != nil {
		return nil, NewProviderError("vertexai", "failed to load Google credentials", err)
	}

	// Use project ID from credentials if not provided
	if projectID == "" {
		projectID = creds.ProjectID
	}

	return &VertexAIProvider{
		projectID:         projectID,
		location:          location,
		publisher:         publisher,
		model:             parsedModel,
		tokenSource:       creds.TokenSource,
		timeout:           timeout,
	}, nil
}

// Name returns the provider name
func (p *VertexAIProvider) Name() string {
	return "vertexai"
}

// Validate checks if the provider is properly configured
func (p *VertexAIProvider) Validate() error {
	if p.projectID == "" {
		return NewProviderError(p.Name(), "project ID is required", nil)
	}
	if p.location == "" {
		return NewProviderError(p.Name(), "location is required", nil)
	}
	if p.publisher == "" {
		return NewProviderError(p.Name(), "publisher is required", nil)
	}
	if p.tokenSource == nil {
		return NewProviderError(p.Name(), "token source is required", nil)
	}

	// Validate location
	validLocations := []string{"us-central1", "us-east4", "us-east5", "europe-west1", "asia-southeast1", "global"}
	isValidLocation := false
	for _, loc := range validLocations {
		if p.location == loc {
			isValidLocation = true
			break
		}
	}
	if !isValidLocation {
		return NewProviderError(p.Name(), fmt.Sprintf("invalid location '%s', must be one of: %v", p.location, validLocations), nil)
	}

	return nil
}

// GenerateCommitMessage generates a commit message using Vertex AI
func (p *VertexAIProvider) GenerateCommitMessage(ctx context.Context, diff string, options *GenerateOptions) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}

	// Get access token
	token, err := p.tokenSource.Token()
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to get access token", err)
	}

	prompt := buildPrompt(diff, options)

	// Build Vertex AI request
	reqBody := vertexAIRequest{
		Contents: []vertexAIContent{
			{
				Role: "user",
				Parts: []vertexAIPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: vertexAIGenerationConfig{
			Temperature:     options.Temperature,
			MaxOutputTokens: options.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to marshal request", err)
	}

	log.Debugln("Vertex AI request payload size:", len(jsonData))

	// Build API endpoint
	endpoint := p.buildEndpoint()
	log.Debugln("Vertex AI endpoint:", endpoint)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", NewProviderError(p.Name(), "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

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
		var errResp vertexAIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return "", NewProviderError(p.Name(), fmt.Sprintf("API error (%d): %s", resp.StatusCode, errResp.Error.Message), nil)
		}
		return "", NewProviderError(p.Name(), fmt.Sprintf("API error (%d): %s", resp.StatusCode, string(body)), nil)
	}

	var apiResp vertexAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", NewProviderError(p.Name(), "failed to parse response", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return "", NewProviderError(p.Name(), "no response from API", nil)
	}

	message := strings.TrimSpace(apiResp.Candidates[0].Content.Parts[0].Text)
	log.Debugln("Vertex AI generated message:", message)

	if apiResp.UsageMetadata.TotalTokenCount > 0 {
		log.Debugln("Tokens used:", apiResp.UsageMetadata.TotalTokenCount)
	}

	return message, nil
}

// buildEndpoint constructs the Vertex AI API endpoint
func (p *VertexAIProvider) buildEndpoint() string {
	// Format: https://{location}-aiplatform.googleapis.com/v1/projects/{project}/locations/{location}/publishers/{publisher}/models/{model}:generateContent
	if p.location == "global" {
		return fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/%s/models/%s:generateContent",
			p.projectID, p.location, p.publisher, p.model)
	}
	return fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/%s/models/%s:generateContent",
		p.location, p.projectID, p.location, p.publisher, p.model)
}

// Vertex AI API request/response structures
type vertexAIRequest struct {
	Contents         []vertexAIContent        `json:"contents"`
	GenerationConfig vertexAIGenerationConfig `json:"generationConfig,omitempty"`
}

type vertexAIContent struct {
	Role  string         `json:"role"`
	Parts []vertexAIPart `json:"parts"`
}

type vertexAIPart struct {
	Text string `json:"text"`
}

type vertexAIGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type vertexAIResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

type vertexAIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
