# Tasks

## 1. Design & Planning

- [x] Define AI provider interface and contract
- [x] Design configuration schema for AI settings
- [x] Create prompt templates for commit message generation
- [x] Define error handling and fallback strategies

## 2. Core Implementation

- [x] Create `ai/` package structure
- [x] Implement base AI provider interface (`ai/types.go`)
- [x] Implement OpenAI provider (`ai/openai.go`)
- [x] Implement Claude/Anthropic provider (`ai/claude.go`)
- [x] Add configuration parsing for AI settings in `helper/config.go`
- [x] Create `cmd/aiCommit.go` command with Cobra integration
- [x] Implement git diff analysis and formatting
- [x] Fix `GetStagedDiff()` to read from git index blobs (not worktree filesystem) — current code reads from `w.Filesystem` which includes unstaged changes
- [x] Implement proper unified diff format with context lines, `@@` hunk headers (replace naive line-by-line comparison)
- [x] Implement `GetWorkingTreeDiff()` for use with `-U` / `--include-unstaged` flag
- [x] Add `-U` / `--include-unstaged` flag to `cmd/aiCommit.go` to toggle between index diff and working-tree diff
- [x] Add interactive prompt for message review/editing
- [x] Integrate with existing commit workflow (reuse from `commit.go`)

## 3. Configuration & Setup

- [x] Update `.workflow.yaml` schema with AI configuration fields
- [x] Add environment variable support for API keys
- [x] Create example configuration in `config/.workflow.yaml.j2`
- [x] Implement configuration validation

## 4. User Experience

- [x] Add spinner/progress indicators during AI processing
- [x] Implement message preview and editing interface
- [x] Add option to regenerate message if unsatisfactory
- [x] Support `--provider` flag to override config
- [x] Support `--no-ai` flag to skip AI and use manual entry
- [x] Support `-s` / `--skip-precommit` flag
- [x] Add `--dry-run` flag to preview without committing

## 5. Error Handling & Resilience

- [x] Handle API authentication failures
- [x] Handle network timeouts and retries
- [x] Handle rate limiting with exponential backoff
- [x] Graceful fallback to manual message entry
- [x] Add logging for debugging AI interactions

## 6. Testing

- [x] Write unit tests for AI provider interface
- [x] Write unit tests for `GetStagedDiff()` verifying index-only behavior (no working-tree leak)
- [x] Write unit tests for `GetWorkingTreeDiff()` (with `-U` flag behavior)
- [x] Write unit tests for unified diff format (context lines, hunk headers)
- [x] Write unit tests for OpenAI provider (with mocks)
- [x] Write unit tests for Claude provider (with mocks)
- [x] Write unit tests for configuration parsing
- [x] Write integration tests for `ai-commit` command (including `-U` flag scenario)
- [x] Manual testing with real API keys
- [x] Test error scenarios (no network, invalid key, etc.)
- [x] Test scenario: staged changes + unstaged modifications on same file (no leak without `-U`)
- [x] Test scenario: `-U` flag includes working-tree changes for staged files only

## 7. Documentation

- [x] Update `README.md` with AI commit feature
- [x] Document configuration options
- [x] Add setup guide for API keys
- [x] Document supported AI providers
- [x] Add usage examples and screenshots
- [x] Document privacy/security considerations
- [x] Update `AGENTS.md` with AI-related commands

## 8. Security & Privacy

- [x] Document data sharing with AI providers
- [x] Add warning about sensitive information in diffs
- [x] Implement option to exclude certain files from AI analysis
- [x] Ensure API keys are not logged or exposed
- [x] Add pre-commit hook to prevent committing API keys

## 9. Polish & Release

- [x] Add bash completion for new command
- [x] Update version number
- [x] Create changelog entry
- [x] Test on multiple platforms (Linux, macOS)
- [x] Performance testing and optimization
