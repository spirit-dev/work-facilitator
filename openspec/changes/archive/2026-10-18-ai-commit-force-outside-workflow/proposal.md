# Proposal

## Why

The `ai-commit` command already accepts a `-f` / `--force-commit` flag, and the pre-run check lets the command proceed when the repo has no current workflow. But the rest of the command assumes workflow data exists: the AI prompt, the commit-standard validation, and the push all read `RootRepo.CurrentWorkflowData`, which is a zero-value struct outside a workflow. The result is an empty branch name in the AI context, a standard check against an empty branch, and a broken push refspec (`refs/heads/:refs/heads/`) that fails. Users who want to commit AI-assisted outside an active workflow currently cannot.

## What Changes

- Make `work-facilitator ai-commit -f` fully functional outside a workflow:
  - Resolve the current git branch (via `RepoHead()`) when no workflow is active, and use it as the AI prompt branch context instead of the empty workflow branch.
  - Skip the workflow pre-message prefix (`CurrentWorkflowData.Commit`) when validating the standard — validation runs against the actual current branch (or no branch when enforcing standards that require one).
  - Push to the current git branch instead of the empty workflow branch.
- No change to behavior when a workflow is active — the normal `ai-commit` path is untouched.
- Assumption (recorded): the requested "commit-ai" refers to the existing `ai-commit` command; no new command or alias is introduced.

## Capabilities

### New Capabilities

- `git-operations`: add the forced AI-assisted commit requirement — `ai-commit -f` SHALL complete commit and push outside an active workflow using the current git branch. (This capability has no archived main spec yet; the `ai-assisted-commit` change carries its delta.)

### Modified Capabilities

<!-- None: no archived main spec exists yet to modify. -->

## Impact

- `src/work-facilitator/cmd/aiCommit.go`: branch resolution and pre-message handling when `forceAICommitArg` is set and `RootRepo.HasCurrentWorkflow` is false.
- Possibly `src/work-facilitator/helper/repo.go`: no new helper needed — `RepoHead()` already exists and returns the current branch reference.
- Tests: `cmd/aiCommit_test.go` for the new branch-resolution/validation paths.
