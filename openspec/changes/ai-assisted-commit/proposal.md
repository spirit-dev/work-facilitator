# ai-assisted-commit

## Why

This feature is a dedicated command that execute a commit, assisted by an AI assistant (OpenAI, Claude, Vertex AI, or other)

## What

A new `ai-commit` command will be added to work-facilitator that leverages AI assistants (OpenAI, Claude, Google Vertex AI, or others) to:

- Analyze staged changes in the git repository
- Generate contextual, meaningful commit messages based on the diff
- Allow user review and editing of the AI-generated message before committing
- Support multiple AI providers with configurable API keys
- Maintain compatibility with existing commit workflow standards and patterns

This extends the current `commit` command functionality by adding intelligent commit message generation.

## How

### Implementation Approach

1. **New Command Structure**
   - Create `aiCommit.go` in `src/work-facilitator/cmd/`
   - Follow existing Cobra command pattern similar to `commit.go`
   - Reuse existing git operations from `helper/repo.go`

2. **AI Provider Integration**
   - Create new package `src/work-facilitator/ai/` for AI provider abstractions
   - Implement provider interface supporting:
     - OpenAI (GPT-4, GPT-3.5)
     - Anthropic Claude (Claude 3.5 Sonnet, Claude 3 Opus)
     - Google Vertex AI (Gemini 2.5 Flash, Gemini 2.5 Pro)
     - Configurable custom providers
   - Use HTTP clients for API communication
   - Service account key authentication for Vertex AI

3. **Configuration**
   - Extend `~/.workflow.yaml` with AI-specific settings:
     - `ai_provider`: Selected provider (openai, claude, vertexai)
     - `ai_api_key`: API key (or reference to environment variable)
     - `ai_model`: Specific model to use
     - `ai_enabled`: Feature toggle
     - `ai_max_tokens`: Response size limit
     - `ai_temperature`: Creativity parameter
     - `ai_google_project_id`: Google Cloud project ID (Vertex AI)
     - `ai_google_location`: GCP region (Vertex AI)
     - `ai_google_service_account_key`: Path to service account key (Vertex AI)

4. **Workflow**
   - Analyze `git diff --staged` to get changes
   - Send diff to AI provider with prompt template
   - Receive and display suggested commit message
   - Allow user to accept, edit, or regenerate
   - Proceed with standard commit flow using final message

5. **Error Handling**
   - Graceful fallback to manual message entry if AI fails
   - API rate limiting handling
   - Network error recovery
   - Invalid API key detection

## Impact

### User Experience

- **Positive**: Faster, more consistent commit messages; reduced cognitive load
- **Learning Curve**: Minimal - optional feature with familiar fallback

### Codebase

- **New Files**: `cmd/aiCommit.go`, `ai/types.go`, `ai/openai.go`, `ai/claude.go`, `ai/vertexai.go`, `ai/types_test.go`, `ai/vertexai_test.go`
- **Modified Files**: `helper/config.go` (new config fields), `common/common.go` (config struct), `README.md` (documentation), `AGENTS.md` (provider list)
- **Dependencies**: New Go modules - `github.com/golang-jwt/jwt/v5` for Vertex AI JWT authentication

### Configuration

- Backward compatible - feature is opt-in
- Existing workflows unaffected
- New configuration section in `~/.workflow.yaml`

### Security

- API keys stored in config or environment variables
- Diff content sent to external AI services (privacy consideration)
- Need to document data sharing implications

### Performance

- Network latency added (1-3 seconds typical)
- Async/timeout handling required
- Minimal local performance impact

### Maintenance

- New external API dependencies to monitor
- Provider API changes may require updates
- Additional test coverage needed
