/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package ai

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
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
	model             string
	serviceAccountKey *ServiceAccountKey
	timeout           time.Duration
	cachedToken       string
	tokenExpiry       time.Time
}

// ServiceAccountKey represents a Google Cloud service account key
type ServiceAccountKey struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
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

	// Load service account key
	key, err := loadServiceAccountKey(serviceAccountKeyPath)
	if err != nil {
		return nil, NewProviderError("vertexai", "failed to load service account key", err)
	}

	// Use project ID from key if not provided
	if projectID == "" {
		projectID = key.ProjectID
	}

	return &VertexAIProvider{
		projectID:         projectID,
		location:          location,
		model:             model,
		serviceAccountKey: key,
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
	if p.serviceAccountKey == nil { // pragma: allowlist secret
		return NewProviderError(p.Name(), "service account key is required", nil)
	}
	if p.serviceAccountKey.PrivateKey == "" {
		return NewProviderError(p.Name(), "service account private key is missing", nil)
	}

	// Validate location
	validLocations := []string{"us-central1", "us-east4", "europe-west1", "asia-southeast1", "global"}
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
	token, err := p.getAccessToken()
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
	req.Header.Set("Authorization", "Bearer "+token)

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
	// Format: https://{location}-aiplatform.googleapis.com/v1/projects/{project}/locations/{location}/publishers/google/models/{model}:generateContent
	if p.location == "global" {
		return fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
			p.projectID, p.location, p.model)
	}
	return fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		p.location, p.projectID, p.location, p.model)
}

// getAccessToken gets or refreshes the OAuth2 access token
func (p *VertexAIProvider) getAccessToken() (string, error) {
	// Check if we have a cached token that's still valid
	if p.cachedToken != "" && time.Now().Before(p.tokenExpiry) {
		log.Debugln("Using cached Vertex AI access token")
		return p.cachedToken, nil
	}

	log.Debugln("Generating new Vertex AI access token")

	// Create JWT
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   p.serviceAccountKey.ClientEmail,
		"sub":   p.serviceAccountKey.ClientEmail,
		"aud":   vertexAITokenURL,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
		"scope": vertexAIScope,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Parse private key
	privateKey, err := parsePrivateKey(p.serviceAccountKey.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Sign JWT
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Exchange JWT for access token
	reqBody := fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s", signedToken)
	resp, err := http.Post(vertexAITokenURL, "application/x-www-form-urlencoded", strings.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to exchange JWT for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Cache the token
	p.cachedToken = tokenResp.AccessToken
	p.tokenExpiry = now.Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // Refresh 60s before expiry

	return p.cachedToken, nil
}

// loadServiceAccountKey loads a service account key from a file
func loadServiceAccountKey(path string) (*ServiceAccountKey, error) {
	// Support environment variable expansion
	if strings.HasPrefix(path, "$") {
		envVar := strings.TrimPrefix(path, "$")
		path = os.Getenv(envVar)
		if path == "" {
			return nil, fmt.Errorf("environment variable %s is not set", envVar)
		}
	}

	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = strings.Replace(path, "~", home, 1)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key file: %w", err)
	}

	var key ServiceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %w", err)
	}

	return &key, nil
}

// parsePrivateKey parses a PEM-encoded RSA private key
func parsePrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
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
