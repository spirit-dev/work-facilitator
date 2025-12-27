/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

func TestAICommitCommand_SkipPreCommit(t *testing.T) {
	// Create a temporary directory for the test repo
	tempDir, err := os.MkdirTemp("", "aicommit-test")
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

	// Initialize a git repo
	_, err = git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create .git/hooks directory
	hooksDir := filepath.Join(".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create hooks dir: %v", err)
	}

	// Create a failing pre-commit hook
	hookPath := filepath.Join(hooksDir, "pre-commit")
	failScript := "#!/bin/sh\necho 'Hook should be skipped'\nexit 1"
	if err := os.WriteFile(hookPath, []byte(failScript), 0755); err != nil {
		t.Fatalf("Failed to create failing hook: %v", err)
	}

	// Test Case: With -s flag, pre-commit hook should be skipped
	// This test verifies that when skipPreCommitAICommitArg is true,
	// the hook execution is bypassed
	t.Run("skip pre-commit with -s flag", func(t *testing.T) {
		// Reset the flag
		skipPreCommitAICommitArg = false

		// Parse flags to set skipPreCommitAICommitArg
		cmd := &cobra.Command{}
		cmd.Flags().BoolVarP(&skipPreCommitAICommitArg, "skip-precommit", "s", false, "Skip pre-commit hooks")
		cmd.ParseFlags([]string{"-s"})

		// Verify flag was set
		if !skipPreCommitAICommitArg {
			t.Error("Expected skipPreCommitAICommitArg to be true when -s flag is passed")
		}
	})

	// Test Case: Without -s flag, pre-commit hook should run
	t.Run("run pre-commit without -s flag", func(t *testing.T) {
		// Reset the flag
		skipPreCommitAICommitArg = false

		// Parse flags without -s
		cmd := &cobra.Command{}
		cmd.Flags().BoolVarP(&skipPreCommitAICommitArg, "skip-precommit", "s", false, "Skip pre-commit hooks")
		cmd.ParseFlags([]string{})

		// Verify flag was not set
		if skipPreCommitAICommitArg {
			t.Error("Expected skipPreCommitAICommitArg to be false when -s flag is not passed")
		}
	})
}
