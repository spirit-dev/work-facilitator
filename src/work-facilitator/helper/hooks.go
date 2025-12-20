package helper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// RunPreCommitHooks executes git pre-commit hooks
// It first tries 'git hook run pre-commit' (git 2.29+)
// If that fails or isn't supported, it falls back to executing .git/hooks/pre-commit directly
func RunPreCommitHooks() error {
	SpinStartDisplay("Running pre-commit hooks")

	// Method 1: Try 'git hook run --ignore-missing pre-commit'
	// This is the preferred modern way as it handles configuration properly
	// --ignore-missing ensures we don't fail if no hook exists
	cmd := exec.Command("git", "hook", "run", "--ignore-missing", "pre-commit")

	// Capture output to show to user if it fails
	output, err := cmd.CombinedOutput()

	if err == nil {
		// Success (hook passed or was missing)
		SpinStopDisplay("success")
		return nil
	}

	// If git hook run failed with a non-zero exit code, it means the hook ran and failed
	// OR git command failed (e.g. not a git repo, command not found)
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
		outputStr := string(output)

		// Check if it's a git error rather than a hook failure
		// Git errors usually start with "git:" or "fatal:"
		isGitError := strings.Contains(outputStr, "git: 'hook' is not a git command") ||
			strings.Contains(outputStr, "fatal: not a git repository")

		if !isGitError {
			// It's likely a hook failure
			SpinStopDisplay("fail")
			log.Warningln("Pre-commit hooks failed:")
			fmt.Println(outputStr)
			return fmt.Errorf("pre-commit hooks failed")
		}
		// If it looks like a git command error, we fall through to fallback
	}

	// Method 2: Execute .git/hooks/pre-commit directly
	// We use repoBasePath() to find the absolute path to the repo root
	// Note: repoBasePath() is available in the helper package and assumes repo is initialized
	gitDir := filepath.Join(repoBasePath(), ".git")
	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")

	// Check if file exists
	info, err := os.Stat(hookPath)
	if os.IsNotExist(err) {
		// No hook exists, so we consider this a success (nothing to run)
		// But if the previous command failed and this one doesn't exist,
		// it implies the previous failure was likely due to 'git hook' not being supported
		// and no manual hook existing.
		SpinStopDisplay("success")
		log.Debugln("No pre-commit hook found, skipping")
		return nil
	}

	if err != nil {
		SpinStopDisplay("fail")
		return fmt.Errorf("failed to check pre-commit hook: %w", err)
	}

	// Check if executable (simplistic check for unix)
	if info.Mode()&0111 == 0 {
		SpinStopDisplay("success")
		log.Debugln("Pre-commit hook exists but is not executable, skipping")
		return nil
	}

	// Execute the hook
	cmd = exec.Command(hookPath)
	output, err = cmd.CombinedOutput()

	if err != nil {
		SpinStopDisplay("fail")
		log.Warningln("Pre-commit hooks failed:")
		fmt.Println(string(output))
		return fmt.Errorf("pre-commit hooks failed: %w", err)
	}

	SpinStopDisplay("success")
	return nil
}
