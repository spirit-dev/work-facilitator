# Tasks

## 1. Design & Planning

- [x] Define AI provider interface and contract
- [ ] Design configuration schema for AI settings
- [ ] Create prompt templates for commit message generation
- [ ] Define error handling and fallback strategies

## 2. Core Implementation

- [x] Create `ai/` package structure
- [x] Implement base AI provider interface (`ai/types.go`)
- [x] Implement OpenAI provider (`ai/openai.go`)
- [x] Implement Claude/Anthropic provider (`ai/claude.go`)
- [x] Add configuration parsing for AI settings in `helper/config.go`
- [x] Create `cmd/aiCommit.go` command with Cobra integration
- [x] Implement git diff analysis and formatting
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
- [ ] Add option to regenerate message if unsatisfactory
- [x] Support `--provider` flag to override config
- [x] Support `--no-ai` flag to skip AI and use manual entry
- [x] Support `-s` / `--skip-precommit` flag
- [x] Add `--dry-run` flag to preview without committing

## 5. Error Handling & Resilience

- [x] Handle API authentication failures
- [ ] Handle network timeouts and retries
- [ ] Handle rate limiting with exponential backoff
- [x] Graceful fallback to manual message entry
- [x] Add logging for debugging AI interactions

## 6. Testing

- [x] Write unit tests for AI provider interface
- [ ] Write unit tests for OpenAI provider (with mocks)
- [ ] Write unit tests for Claude provider (with mocks)
- [ ] Write unit tests for configuration parsing
- [ ] Write integration tests for `ai-commit` command
- [ ] Manual testing with real API keys
- [ ] Test error scenarios (no network, invalid key, etc.)

## 7. Documentation

- [x] Update `README.md` with AI commit feature
- [ ] Document configuration options
- [ ] Add setup guide for API keys
- [ ] Document supported AI providers
- [ ] Add usage examples and screenshots
- [ ] Document privacy/security considerations
- [x] Update `AGENTS.md` with AI-related commands

## 8. Security & Privacy

- [ ] Document data sharing with AI providers
- [ ] Add warning about sensitive information in diffs
- [ ] Implement option to exclude certain files from AI analysis
- [ ] Ensure API keys are not logged or exposed
- [ ] Add pre-commit hook to prevent committing API keys

## 9. Polish & Release

- [ ] Add bash completion for new command
- [ ] Update version number
- [ ] Create changelog entry
- [ ] Test on multiple platforms (Linux, macOS)
- [ ] Performance testing and optimization
