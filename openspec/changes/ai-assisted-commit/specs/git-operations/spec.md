# Git-Operations Specification

## Purpose

This specification defines the requirements for AI-assisted commit functionality within the work-facilitator tool. The feature enables developers to generate intelligent, contextual commit messages using AI providers while maintaining compatibility with existing git workflows and commit standards.

## Requirements

### Requirement: AI-Assisted Commit Message Generation

**User Story:** As a developer, I want to use AI to generate meaningful commit messages based on my staged changes, so that I can save time and maintain consistent, high-quality commit history.

#### Acceptance Criteria

1. WHEN user runs `work-facilitator ai-commit` THEN system SHALL analyze staged git changes
2. WHEN staged changes exist THEN system SHALL send diff to configured AI provider
3. WHEN AI provider responds THEN system SHALL display suggested commit message to user
4. WHEN user reviews message THEN system SHALL provide options to accept, edit, or regenerate
5. WHEN user accepts message THEN system SHALL proceed with standard commit workflow
6. IF AI provider fails THEN system SHALL fallback to manual message entry
7. IF no changes are staged THEN system SHALL prompt user to stage changes first
8. WHEN commit message is generated THEN system SHALL respect existing commit standards and patterns

#### Scenario: Happy Path - AI-Generated Commit

1. Developer stages changes with `git add`
2. Developer runs `work-facilitator ai-commit`
3. System displays "Analyzing changes..." spinner
4. System sends git diff to AI provider (e.g., OpenAI)
5. AI provider returns suggested message: "feat: add user authentication middleware"
6. System displays message and prompts: "[A]ccept, [E]dit, [R]egenerate, [C]ancel?"
7. Developer presses 'A' to accept
8. System commits with AI-generated message
9. System pushes to remote (unless --no-push flag used)
10. System displays success message

#### Scenario: Edit AI-Generated Message

1. Developer runs `work-facilitator ai-commit`
2. System generates message: "update config file"
3. Developer selects 'E' to edit
4. System opens editor with pre-filled message
5. Developer modifies to: "fix: correct API endpoint in config"
6. Developer saves and exits editor
7. System commits with edited message

#### Scenario: AI Provider Failure

1. Developer runs `work-facilitator ai-commit`
2. System attempts to contact AI provider
3. Network timeout occurs
4. System displays warning: "AI provider unavailable, falling back to manual entry"
5. System prompts for manual commit message
6. Developer enters message manually
7. System proceeds with normal commit workflow

### Requirement: Staged-Only Diff Guarantee

**User Story:** As a developer, I want the AI to generate commit messages based ONLY on my staged changes, so that unstaged modifications never produce misleading commit messages.

#### Acceptance Criteria

1. WHEN user runs `work-facilitator ai-commit` THEN system SHALL read staged file content from git index blobs (not working tree)
2. WHEN a file has both staged and unstaged modifications THEN system SHALL send only the staged version to the AI
3. WHEN `-U` / `--include-unstaged` flag is provided THEN system SHALL read working-tree content for staged files instead of index blobs
4. WHEN `-U` flag is provided THEN system SHALL still only include files that have staged entries in the index
5. WHEN unstaged changes exist on files NOT in the index THEN system SHALL exclude those files regardless of the `-U` flag
6. WHEN diff is generated THEN system SHALL produce proper unified diff format with context lines and `@@` hunk headers
7. WHEN untracked files exist THEN system SHALL exclude them from the diff (unless staged via `--all-files`)

#### Scenario: Unstaged Changes Do Not Leak

1. Developer edits `main.go` and stages with `git add main.go`
2. Developer edits `main.go` again (second change, unstaged)
3. Developer runs `work-facilitator ai-commit`
4. System reads staged blob for `main.go` from git object store (not working directory)
5. System sends diff containing only the staged change to the AI
6. AI generates message based only on the staged change
7. The unstaged modification is NOT reflected in the commit message

#### Scenario: Include Unstaged with -U Flag

1. Developer edits `main.go` and stages with `git add main.go`
2. Developer edits `main.go` again (second change, unstaged)
3. Developer also edits `utils.go` without staging it
4. Developer runs `work-facilitator ai-commit -U`
5. System reads working-tree content for `main.go` (includes unstaged changes)
6. System does NOT include `utils.go` in the diff (not staged)
7. AI generates message based on full `main.go` changes, no `utils.go` context

#### Scenario: Multiple Files, Partial Staging

1. Developer stages `api.go` and `model.go` (both have staged-only changes)
2. Developer has unstaged modifications on `config.go` (not staged at all)
3. Developer runs `work-facilitator ai-commit`
4. System generates diff for `api.go` and `model.go` from index blobs
5. System excludes `config.go` entirely
6. AI generates message based on `api.go` and `model.go` only

### Requirement: Multi-Provider Support

**User Story:** As a developer, I want to choose between different AI providers (OpenAI, Claude, etc.), so that I can use my preferred service or switch if one is unavailable.

#### Acceptance Criteria

1. WHEN user configures `ai_provider: openai` THEN system SHALL use OpenAI API
2. WHEN user configures `ai_provider: claude` THEN system SHALL use Anthropic Claude API
3. WHEN user provides `--provider` flag THEN system SHALL override config setting
4. IF configured provider is invalid THEN system SHALL display error and list valid providers
5. WHEN provider requires API key THEN system SHALL read from config or environment variable

#### Scenario: Switch Provider via Flag

1. Developer has `ai_provider: openai` in config
2. Developer runs `work-facilitator ai-commit --provider claude`
3. System uses Claude API instead of OpenAI
4. System generates commit message using Claude
5. Workflow proceeds normally

### Requirement: Configuration Management

**User Story:** As a developer, I want to configure AI settings in my workflow config file, so that I can customize the AI behavior to my preferences.

#### Acceptance Criteria

1. WHEN user adds AI config to `~/.workflow.yaml` THEN system SHALL parse and validate settings
2. WHEN `ai_enabled: false` THEN system SHALL disable AI features
3. WHEN `ai_api_key` is set THEN system SHALL use provided key
4. WHEN `ai_api_key` starts with `$` THEN system SHALL read from environment variable
5. WHEN `ai_model` is specified THEN system SHALL use that model
6. IF required config is missing THEN system SHALL display helpful error message

#### Scenario: Environment Variable API Key

1. Developer sets `ai_api_key: $OPENAI_API_KEY` in config
2. Developer exports `OPENAI_API_KEY=sk-...` in shell
3. Developer runs `work-facilitator ai-commit`
4. System reads API key from environment variable
5. System successfully authenticates with OpenAI

### Requirement: Privacy and Security

**User Story:** As a developer, I want to understand what data is sent to AI providers and have control over it, so that I can protect sensitive information.

#### Acceptance Criteria

1. WHEN AI feature is first used THEN system SHALL display privacy notice about data sharing
2. WHEN diff contains sensitive patterns THEN system SHALL warn user before sending
3. WHEN user configures `ai_exclude_patterns` THEN system SHALL filter matching files from diff
4. WHEN API key is invalid THEN system SHALL display error without logging the key
5. IF user cancels AI operation THEN system SHALL not send any data to provider

#### Scenario: Sensitive File Exclusion

1. Developer configures `ai_exclude_patterns: ["*.env", "secrets/*"]`
2. Developer stages changes including `.env` file
3. Developer runs `work-facilitator ai-commit`
4. System filters `.env` from diff sent to AI
5. AI generates message based on remaining changes only

### Requirement: Standard Compliance

**User Story:** As a developer, I want AI-generated messages to comply with my team's commit standards, so that I maintain consistency across the repository.

#### Acceptance Criteria

1. WHEN commit standard is enforced THEN system SHALL validate AI-generated message
2. IF AI message violates standard THEN system SHALL prompt AI to regenerate with standard context
3. WHEN branch pattern exists THEN system SHALL include branch context in AI prompt
4. WHEN commit prefix is required THEN system SHALL ensure AI includes it

#### Scenario: Standard Enforcement

1. Developer has `enforce_standard: true` and `commit_expr: "^(feat|fix|docs):.*"`
2. Developer runs `work-facilitator ai-commit`
3. AI generates: "add new feature"
4. System detects standard violation
5. System regenerates with prompt: "Use conventional commit format (feat|fix|docs)"
6. AI returns: "feat: add new feature"
7. System validates and accepts message
