package helper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestRunPreCommitHooks(t *testing.T) {
	// Create a temporary directory for the test repo
	tempDir, err := os.MkdirTemp("", "hooks-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current wd and change to temp dir
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current wd: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Initialize a git repo so 'git hook run' can work (if git is installed)
	// and so repoBasePath() works
	r, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Set the global repo variable for repoBasePath() to work
	// We save the old one to restore it (though it's likely nil in test context)
	oldRepo := repo
	repo = r
	defer func() { repo = oldRepo }()

	// Create .git/hooks directory
	hooksDir := filepath.Join(".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Test Case 1: No hooks exist
	// Should return nil (success)
	if err := RunPreCommitHooks(); err != nil {
		t.Errorf("Expected success when no hooks exist, got error: %v", err)
	}

	// Test Case 2: Hook exists and succeeds
	hookPath := filepath.Join(hooksDir, "pre-commit")
	// Create a simple success hook
	successScript := "#!/bin/sh\nexit 0"
	if err := os.WriteFile(hookPath, []byte(successScript), 0755); err != nil {
		t.Fatalf("Failed to create success hook: %v", err)
	}

	if err := RunPreCommitHooks(); err != nil {
		t.Errorf("Expected success when hook succeeds, got error: %v", err)
	}

	// Test Case 3: Hook exists and fails
	// Create a failing hook
	failScript := "#!/bin/sh\necho 'Hook failed'\nexit 1"
	if err := os.WriteFile(hookPath, []byte(failScript), 0755); err != nil {
		t.Fatalf("Failed to create failing hook: %v", err)
	}

	if err := RunPreCommitHooks(); err == nil {
		t.Error("Expected error when hook fails, got nil")
	}
}
