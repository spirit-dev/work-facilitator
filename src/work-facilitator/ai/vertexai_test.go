/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewVertexAIProvider(t *testing.T) {
	// Create a temporary service account key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")

	keyContent := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "",
		"private_key": "",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com"
	}`

	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	tests := []struct {
		name          string
		projectID     string
		location      string
		keyPath       string
		model         string
		wantErr       bool
		wantProjectID string
	}{
		{
			name:          "Valid configuration with explicit project ID",
			projectID:     "my-project",
			location:      "us-central1",
			keyPath:       keyPath,
			model:         "gemini-2.5-flash",
			wantErr:       false,
			wantProjectID: "my-project",
		},
		{
			name:          "Valid configuration with project ID from key",
			projectID:     "",
			location:      "us-central1",
			keyPath:       keyPath,
			model:         "gemini-2.5-flash",
			wantErr:       false,
			wantProjectID: "test-project",
		},
		{
			name:      "Invalid key path",
			projectID: "my-project",
			location:  "us-central1",
			keyPath:   "/nonexistent/path.json",
			model:     "gemini-2.5-flash",
			wantErr:   true,
		},
		{
			name:          "Default model",
			projectID:     "my-project",
			location:      "us-central1",
			keyPath:       keyPath,
			model:         "",
			wantErr:       false,
			wantProjectID: "my-project",
		},
		{
			name:          "Default location",
			projectID:     "my-project",
			location:      "",
			keyPath:       keyPath,
			model:         "gemini-2.5-flash",
			wantErr:       false,
			wantProjectID: "my-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewVertexAIProvider(tt.projectID, tt.location, tt.keyPath, tt.model, 0)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewVertexAIProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if provider == nil {
					t.Error("Expected provider to be non-nil")
					return
				}
				if provider.projectID != tt.wantProjectID {
					t.Errorf("Expected project ID %s, got %s", tt.wantProjectID, provider.projectID)
				}
				if provider.Name() != "vertexai" {
					t.Errorf("Expected provider name 'vertexai', got '%s'", provider.Name())
				}
			}
		})
	}
}

func TestVertexAIProviderValidate(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")

	keyContent := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "",
		"private_key": "",
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "123456789",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com"
	}`

	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	tests := []struct {
		name      string
		projectID string
		location  string
		keyPath   string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid configuration",
			projectID: "test-project",
			location:  "us-central1",
			keyPath:   keyPath,
			wantErr:   false,
		},
		{
			name:      "Project ID from key file",
			projectID: "",
			location:  "us-central1",
			keyPath:   keyPath,
			wantErr:   false, // Project ID is extracted from key file
		},
		{
			name:      "Invalid location",
			projectID: "test-project",
			location:  "invalid-location",
			keyPath:   keyPath,
			wantErr:   true,
			errMsg:    "invalid location",
		},
		{
			name:      "Valid global location",
			projectID: "test-project",
			location:  "global",
			keyPath:   keyPath,
			wantErr:   false,
		},
		{
			name:      "Valid europe-west1 location",
			projectID: "test-project",
			location:  "europe-west1",
			keyPath:   keyPath,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewVertexAIProvider(tt.projectID, tt.location, tt.keyPath, "", 0)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			err = provider.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("Expected error message containing '%s', got nil", tt.errMsg)
				}
			}
		})
	}
}

func TestBuildEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		location  string
		model     string
		want      string
	}{
		{
			name:      "US Central 1",
			projectID: "my-project",
			location:  "us-central1",
			model:     "gemini-2.5-flash",
			want:      "https://us-central1-aiplatform.googleapis.com/v1/projects/my-project/locations/us-central1/publishers/google/models/gemini-2.5-flash:generateContent",
		},
		{
			name:      "Global location",
			projectID: "my-project",
			location:  "global",
			model:     "gemini-2.5-flash",
			want:      "https://aiplatform.googleapis.com/v1/projects/my-project/locations/global/publishers/google/models/gemini-2.5-flash:generateContent",
		},
		{
			name:      "Europe West 1",
			projectID: "my-project",
			location:  "europe-west1",
			model:     "gemini-2.5-pro",
			want:      "https://europe-west1-aiplatform.googleapis.com/v1/projects/my-project/locations/europe-west1/publishers/google/models/gemini-2.5-pro:generateContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &VertexAIProvider{
				projectID: tt.projectID,
				location:  tt.location,
				model:     tt.model,
			}

			got := provider.buildEndpoint()
			if got != tt.want {
				t.Errorf("buildEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadServiceAccountKey(t *testing.T) {
	tmpDir := t.TempDir()

	validKeyContent := `{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "",
		"private_key": "",
		"client_email": "test@test-project.iam.gserviceaccount.com"
	}`

	validKeyPath := filepath.Join(tmpDir, "valid-key.json")
	if err := os.WriteFile(validKeyPath, []byte(validKeyContent), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Valid key file",
			path:    validKeyPath,
			wantErr: false,
		},
		{
			name:    "Nonexistent file",
			path:    "/nonexistent/path.json",
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			path:    filepath.Join(tmpDir, "invalid.json"),
			wantErr: true,
		},
	}

	// Create invalid JSON file
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not json"), 0600); err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := loadServiceAccountKey(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("loadServiceAccountKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && key == nil {
				t.Error("Expected key to be non-nil")
			}

			if !tt.wantErr && key.ProjectID != "test-project" {
				t.Errorf("Expected project ID 'test-project', got '%s'", key.ProjectID)
			}
		})
	}
}
