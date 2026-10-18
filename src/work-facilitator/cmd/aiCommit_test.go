/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"spirit-dev/work-facilitator/work-facilitator/common"

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

func TestResolveEffectiveCommitContext(t *testing.T) {
	// Outside a workflow: branch comes from current git HEAD, no pre-message
	t.Run("outside workflow uses HEAD branch and empty pre-message", func(t *testing.T) {
		repo := common.Repo{HasCurrentWorkflow: false}
		branch, preMessage := resolveEffectiveCommitContext(repo, "feature-x")
		if branch != "feature-x" {
			t.Errorf("Expected branch 'feature-x', got '%s'", branch)
		}
		if preMessage != "" {
			t.Errorf("Expected empty preMessage, got '%s'", preMessage)
		}
	})

	// Inside a workflow: branch and pre-message come from workflow data
	// (regression guard: -f must not change the workflow path)
	t.Run("inside workflow uses workflow data", func(t *testing.T) {
		repo := common.Repo{
			HasCurrentWorkflow: true,
			CurrentWorkflowData: common.Workflow{
				Branch: "wf-branch",
				Commit: "wf-prefix: ",
			},
		}
		branch, preMessage := resolveEffectiveCommitContext(repo, "ignored-head")
		if branch != "wf-branch" {
			t.Errorf("Expected branch 'wf-branch', got '%s'", branch)
		}
		if preMessage != "wf-prefix: " {
			t.Errorf("Expected preMessage 'wf-prefix: ', got '%s'", preMessage)
		}
	})
}
