package helper

import (
	"os"
	"path/filepath"
	"testing"
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

	// Create .git/hooks directory
	hooksDir := filepath.Join(".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Test Case 1: No hooks exist
	// Should pass (return nil)
	if err := RunPreCommitHooks(); err != nil {
		t.Errorf("Expected no error when no hooks exist, got: %v", err)
	}

	// Test Case 2: Hook exists and passes
	hookPath := filepath.Join(hooksDir, "pre-commit")
	passScript := "#!/bin/sh\nexit 0"
	if err := os.WriteFile(hookPath, []byte(passScript), 0755); err != nil {
		t.Fatalf("Failed to write pass hook: %v", err)
	}

	// Note: We can't easily test 'git hook run' without a real git repo,
	// but our fallback logic checks for the file execution.
	// To ensure fallback logic is tested, we rely on 'git hook run' failing
	// (which it will, as this isn't a real git repo, just a dir with .git/hooks)
	// or being unavailable.

	if err := RunPreCommitHooks(); err != nil {
		t.Errorf("Expected no error when hook passes, got: %v", err)
	}

	// Test Case 3: Hook exists and fails
	failScript := "#!/bin/sh\necho 'Hook failed'\nexit 1"
	if err := os.WriteFile(hookPath, []byte(failScript), 0755); err != nil {
		t.Fatalf("Failed to write fail hook: %v", err)
	}

	if err := RunPreCommitHooks(); err == nil {
		t.Error("Expected error when hook fails, got nil")
	}
}
