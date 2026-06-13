## 3.8.0

### New Features
- **AI-Assisted Commits**: New `ai-commit` command that generates meaningful commit messages using AI providers
  - Support for OpenAI, Claude/Anthropic, LlamaCPP, and Google VertexAI providers
  - Staged-only diff by default (reads from git index blobs to prevent unstaged changes from leaking into commit messages)
  - `-U`/`--include-unstaged` flag includes working-tree modifications on staged files
  - Interactive review with [A]ccept, [E]dit, [R]egenerate, [C]ancel options
  - Proper unified diff format with `@@` hunk headers and context lines
  - Automatic retry with exponential backoff for network/rate-limit errors
  - Fallback to manual message entry on AI provider failure
  - Privacy controls: sensitive file exclusion patterns, API key protection, data usage warnings
  - Model name displayed alongside context size during generation

### Configuration
- New `ai` section in `~/.workflow.yaml`:
  - `ai.enabled`: Enable/disable AI features
  - `ai.provider`: AI provider (openai, claude, llamacpp, vertexai)
  - `ai.api_key`: API key (supports `$ENV_VAR` references)
  - `ai.model`: Model name (provider-specific defaults)
  - `ai.max_tokens`: Maximum response tokens (default: 1024)
  - `ai.temperature`: Response creativity (default: 0.7)
  - `ai.timeout`: API request timeout in seconds (default: 30)
  - `ai.exclude_patterns`: File patterns to exclude from AI analysis

### Testing
- Unit tests for OpenAI and Claude providers with mock HTTP servers
- Configuration helper tests (TestStandard, CleanString, CleanGlabString, Template, DefineCommit)
- Comprehensive unified diff format tests (17 test cases)
- Staged/working-tree diff isolation tests
