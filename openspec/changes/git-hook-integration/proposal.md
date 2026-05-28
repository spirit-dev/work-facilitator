# git-hook-integration

## Why

Currently, the `work-facilitator` commit commands (`commit` and `ai-commit`) bypass standard git hooks because they use `go-git` which does not natively support hook execution. This means quality checks (linting, tests, etc.) configured in pre-commit hooks are skipped, potentially allowing bad code to be committed.

## What

Integrate pre-commit hook execution into the `commit` and `ai-commit` commands. The integration will:
1.  Attempt to run hooks using `git hook run pre-commit` (standard git method).
2.  Fall back to executing `.git/hooks/pre-commit` directly if the git command fails or isn't available.
3.  Abort the commit process immediately if hooks fail.
4.  Apply to both manual `commit` and AI-assisted `ai-commit` workflows.

## How

1.  **New Helper Module**: Create `src/work-facilitator/helper/hooks.go` to handle hook execution logic.
    *   Function `RunPreCommitHooks()` will implement the hybrid execution strategy.
    *   It will check for `git` binary availability.
    *   It will capture and display stdout/stderr from the hooks.

2.  **Command Integration**:
    *   Modify `src/work-facilitator/cmd/commit.go`: Call `RunPreCommitHooks()` before `RepoCommit`.
    *   Modify `src/work-facilitator/cmd/aiCommit.go`: Call `RunPreCommitHooks()` before `generateAICommitMessage` (to save AI tokens/time if hooks fail).
    *   Add `-s` / `--skip-precommit` flag to both commands to bypass hook execution when needed.

## Impact

*   **Reliability**: Ensures all commits made via `work-facilitator` adhere to project standards defined in hooks.
*   **Workflow**: Users will see hook output in the CLI. Failed hooks will stop the process, requiring users to fix issues before retrying.
*   **Performance**: Slight delay before commit/generation while hooks run.
