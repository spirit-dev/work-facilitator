# Tasks

## 1. Implementation

- [x] Create `src/work-facilitator/helper/hooks.go` with `RunPreCommitHooks` function implementing the hybrid execution strategy (git command + fallback)
- [x] Integrate hook execution into `src/work-facilitator/cmd/commit.go` (manual commit flow)
- [x] Integrate hook execution into `src/work-facilitator/cmd/aiCommit.go` (AI commit flow)
- [x] Implement `-s` / `--skip-precommit` flag for both commands

## 2. Testing

- [x] Verify `commit` command runs hooks and aborts on failure
- [x] Verify `ai-commit` command runs hooks and aborts on failure
- [x] Verify behavior when no hooks are present (should proceed silently)
- [x] Verify output display is clear to the user
- [x] Verify `-s` flag bypasses hook execution

## 3. Documentation

- [x] Update any relevant developer documentation regarding hook support
- [x] Update README with `-s` flag documentation
