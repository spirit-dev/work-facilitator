# work-facilitator

> Take care of it
>
> I love this tool

[![GitLab Sync](https://img.shields.io/badge/gitlab_sync-work_facilitator-blue?style=for-the-badge&logo=gitlab)](https://gitlab-internal.spirit-dev.net/github-mirror/scripts-work-facilitator) <!-- markdownlint-disable MD041 -->
[![GitHub Mirror](https://img.shields.io/badge/github_mirror-work_facilitator-blue?style=for-the-badge&logo=github)](https://github.com/spirit-dev/work-facilitator)

<!--TOC-->

- [Presentation](#presentation)
- [Features](#features)
  - [version](#version)
  - [init](#init)
  - [commit](#commit)
  - [ai-commit](#ai-commit)
    - [Features (AI)](#features-ai)
    - [Configuration](#configuration)
    - [Usage](#usage)
    - [Workflow](#workflow)
    - [Environment Variables](#environment-variables)
    - [Privacy & Security](#privacy--security)
    - [Flags](#flags)
  - [end](#end)
  - [list](#list)
  - [open](#open)
  - [pause](#pause)
  - [status](#status)
  - [use](#use)
  - [initLazy](#initlazy)
  - [completion](#completion)

<!--TOC-->

## Presentation

This our main script. It will manage our interaction between JIRA - GitHub - Gitlab

It will also enforce some standards when it comes to branch and message conventions

## Features

Bellow some features described

### version

Display the current version of the script

### init

Initialize a work in link with JIRA or GitLab

Also, it takes the standards in consideration

### commit

Commit current changes properly prefixed

**Pre-commit Hooks**: The `commit` command automatically executes git pre-commit hooks (if configured). It supports both standard git hooks (`.git/hooks/pre-commit`) and the `pre-commit` framework (via `git hook run`). If hooks fail, the commit is aborted.

### ai-commit

**AI-Assisted Commit Message Generation** - Leverage AI to generate meaningful commit messages based on your staged changes.

**Pre-commit Hooks**: Like the standard commit command, `ai-commit` runs pre-commit hooks before generating the message. This ensures you don't waste AI resources on code that fails validation.

The `ai-commit` command analyzes your git diff and uses AI providers (OpenAI, Claude, or Vertex AI) to generate contextual, well-formatted commit messages that follow your project's standards.

#### Features (AI)

- **Multiple AI Providers**: Support for OpenAI (GPT-4), Anthropic Claude, and Google Vertex AI (Gemini)
- **Interactive Review**: Review, edit, or regenerate AI-generated messages before committing
- **Standard Compliance**: Automatically follows configured commit message standards
- **Privacy Controls**: Exclude sensitive files from AI analysis
- **Fallback Support**: Gracefully falls back to manual entry if AI fails

#### Configuration

Add AI settings to your `~/.workflow.yaml`:

```yaml
ai:
  enabled: true
  provider: "openai"  # Options: openai, claude, vertexai
  api_key: "$OPENAI_API_KEY"  # Use $ENV_VAR to reference environment variables
  model: ""  # Leave empty for default (gpt-4 for openai, claude-3-5-sonnet-20241022 for claude, gemini-2.5-flash for vertexai)
  max_tokens: 1024
  temperature: 0.7
  timeout: 30  # seconds
  exclude_patterns: []  # File patterns to exclude from AI analysis (e.g., ["*.env", "secrets/*"])

  # Vertex AI specific settings (only needed if provider is "vertexai")
  google_project_id: ""  # Google Cloud project ID
  google_location: "us-central1"  # Options: us-central1, us-east4, europe-west1, asia-southeast1, global
  google_service_account_key: ""  # Path to service account key JSON file
```

#### Usage

```bash
# Basic usage - AI generates commit message
work-facilitator ai-commit

# Stage all files and commit with AI
work-facilitator ai-commit -a

# Commit without pushing
work-facilitator ai-commit -n

# Override AI provider
work-facilitator ai-commit --provider claude
work-facilitator ai-commit --provider vertexai

# Skip AI and enter message manually
work-facilitator ai-commit --no-ai

# Preview message without committing
work-facilitator ai-commit -d

# Force commit outside workflow
work-facilitator ai-commit -f
```

#### Workflow

1. Stage your changes with `git add`
2. Run `work-facilitator ai-commit`
3. AI analyzes the diff and generates a commit message
4. Review the message and choose:
   - **[A]ccept**: Use the AI-generated message
   - **[E]dit**: Open in your editor to modify
   - **[R]egenerate**: Request a new message (future feature)
   - **[C]ancel**: Abort the commit
5. Message is validated against commit standards
6. Changes are committed and pushed (unless `-n` flag used)

#### Environment Variables

Set your API key as an environment variable for security:

```bash
# For OpenAI
export OPENAI_API_KEY="sk-..." # pragma: allowlist secret

# For Claude
export ANTHROPIC_API_KEY="sk-ant-..." # pragma: allowlist secret

# For Vertex AI (service account key path)
export GOOGLE_SERVICE_ACCOUNT_KEY="/path/to/service-account-key.json"
```

Then reference it in config:

```yaml
ai:
  api_key: "$OPENAI_API_KEY"  # or "$ANTHROPIC_API_KEY"
  # For Vertex AI:
  google_service_account_key: "$GOOGLE_SERVICE_ACCOUNT_KEY"
```

**Vertex AI Setup:**

For Vertex AI, you need a Google Cloud service account key:

1. Create a service account in Google Cloud Console
2. Grant it the "Vertex AI User" role
3. Download the JSON key file
4. Set the path in your config or environment variable

Example Vertex AI configuration:

```yaml
ai:
  enabled: true
  provider: "vertexai"
  google_project_id: "my-gcp-project"
  google_location: "us-central1"
  google_service_account_key: "/home/user/.config/gcp/service-account-key.json"
  model: "gemini-2.5-flash"
  max_tokens: 1024
  temperature: 0.7
```

#### Privacy & Security

- **Data Sharing**: Git diffs are sent to external AI providers
- **Sensitive Files**: Use `exclude_patterns` to prevent sensitive files from being analyzed
- **API Keys**: Store in environment variables, never commit to repository
- **Local Processing**: All git operations remain local; only diffs are sent to AI

#### Flags

- `-a, --all-files`: Stage all modified files before commit
- `-n, --no-push`: Commit without pushing to remote
- `-f, --force-commit`: Force commit even if not in a workflow
- `-p, --provider <name>`: Override AI provider (openai, claude)
- `--no-ai`: Skip AI generation and enter message manually
- `-d, --dry-run`: Preview commit message without committing

### end

Close a work

**Uncommitted Files Detection**: The `end` command can detect uncommitted or unstaged files before closing a workflow. The detection respects `.gitignore` patterns (both repository-level and global). Configure behavior in `~/.workflow.yaml`:

```yaml
global:
  uncommitted_files_detection: fatal  # Options: disabled, warning, fatal, interactive
```

- `disabled`: Skip the check
- `warning`: Show warning but continue
- `fatal`: Abort if uncommitted files found (default)
- `interactive`: Prompt for confirmation

**Force Flag**: Use `-f` or `--force` to skip the uncommitted files check:

```bash
work-facilitator end -f
work-facilitator end --force -w my-branch
```

**Note**: Files matching `.gitignore` patterns are automatically excluded from detection.

### list

List created works

### open

Open browser directly to the repository

### pause

Pause current work

**Uncommitted Files Detection**: The `pause` command can detect uncommitted or unstaged files before pausing a workflow. The detection respects `.gitignore` patterns (both repository-level and global). Configure behavior in `~/.workflow.yaml`:

```yaml
global:
  uncommitted_files_detection: fatal  # Options: disabled, warning, fatal, interactive
```

- `disabled`: Skip the check
- `warning`: Show warning but continue
- `fatal`: Abort if uncommitted files found (default)
- `interactive`: Prompt for confirmation

**Force Flag**: Use `-f` or `--force` to skip the uncommitted files check:

```bash
work-facilitator pause -f
work-facilitator pause --force -m
```

**Note**: Files matching `.gitignore` patterns are automatically excluded from detection.

### status

Status of the current work

### use

Open a paused work

### initLazy

Create work based on JIRA or Gitlab informations

### completion

Generate completion for Linux / Mac systems
