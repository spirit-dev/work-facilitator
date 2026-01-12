# Agent Guidelines for work-facilitator

This document provides essential guidelines for AI agents working on the `work-facilitator` codebase. Follow these instructions to ensure consistency, quality, and adherence to project standards.

## 1. Build, Lint, and Test Commands

The project uses a `Makefile` for common operations. Always run tests before submitting changes.

### Build

- **Build Standalone Binary**:

  ```bash
  make build-stand-alone
  ```

  Outputs the binary to `dist/work-facilitator`.

- **Run in Development Mode**:

  ```bash
  make run-dev CMD_OPT="<args>"
  # OR directly with go run
  cd src/work-facilitator && go run . <args>
  ```

### Test

- **Run All Tests**:

  ```bash
  make test-dev
  # OR
  cd src/work-facilitator && go test -v ./...
  ```

- **Run a Single Test**:
  To run a specific test function (e.g., `TestNewLlamaCPPProvider`):

  ```bash
  cd src/work-facilitator && go test -v -run TestNewLlamaCPPProvider ./path/to/package
  # Example:
  cd src/work-facilitator && go test -v -run TestNewLlamaCPPProvider ./ai
  ```

### Lint & Quality Checks

- **Pre-commit Hooks**: The project enforces quality via pre-commit hooks including:

  - `hadolint` (Dockerfile linting)
  - `markdownlint` (Documentation)
  - `gitleaks` (Secret detection)
  - `detect-secrets`
  - `gofmt` (Go formatting)

  Ensure your code passes these checks. If you must bypass them (rarely), use the `-s` flag with the commit command, but prefer fixing the issues.

- **Dependency Management**:

  ```bash
  make packages
  # OR
  cd src/work-facilitator && go mod tidy
  ```

## 2. Code Style & Conventions

### General

- **Language**: Go 1.23+
- **Module Name**: `spirit-dev/work-facilitator`
- **Copyright**: All files must start with the copyright header:

  ```go
  /*
  Copyright © 2024 Jean Bordat bordat.jean@gmail.com
  */
  ```

### Project Structure

- `src/work-facilitator/cmd/`: CLI commands (Cobra framework).
- `src/work-facilitator/helper/`: Helper functions and utilities.
- `src/work-facilitator/common/`: Common types and shared constants.
- `src/work-facilitator/ticketing/`: Integrations with ticketing systems (Jira, GitLab).
- `src/work-facilitator/ai/`: AI provider implementations (OpenAI, Claude, VertexAI, LlamaCPP).
- `src/work-facilitator/main.go`: Entry point.

### Naming Conventions

- **Exported Symbols**: Use `CamelCase` (e.g., `NewConfig`, `GenerateCommitMessage`).
- **Unexported Symbols**: Use `camelCase` (e.g., `buildPrompt`, `validateConfig`).
- **Package Names**: Lowercase, single-word (e.g., `ai`, `helper`, `cmd`).
- **Interfaces**: Named with `-er` suffix when possible (e.g., `Provider`, `TicketingSystem`).

### Imports

Group imports in the following order:

1. Standard library (e.g., `fmt`, `os`).
2. External packages (e.g., `github.com/spf13/cobra`).
3. Internal packages (e.g., `spirit-dev/work-facilitator/...`).

Use aliases for common packages to avoid conflicts or improve readability:

```go
import (
    "fmt"

    log "github.com/sirupsen/logrus"
    c "spirit-dev/work-facilitator/work-facilitator/common"
)
```

### Error Handling

- **Explicit Checks**: Always check errors. Do not ignore them using `_`.
- **Logging**:
  - Fatal errors: `log.Fatalln(err)` (terminates execution).
  - Warnings: `log.Warningln("message")`.
  - Debug: `log.Debugln("message")`.
- **Return Values**: Return errors to the caller when possible rather than exiting deep in the call stack, unless it's a CLI command execution flow.

### Configuration

- Use `viper` for configuration management.
- Configuration is loaded from `~/.workflow.yaml`.
- Access config via `helper.NewConfig()` or the global `RootConfig` in `cmd` package.
- When adding new config options:
  1. Update `common.Config` struct.
  2. Update `helper/config.go` to read the value.
  3. Update `README.md` to document the new option.

### AI Providers

- Implement the `ai.Provider` interface.
- Place new providers in `src/work-facilitator/ai/`.
- Ensure `Validate()` method checks for required configuration.
- Use `buildPrompt` helper for consistency across providers.
- Handle timeouts and context cancellation properly.

### Testing

- **File Naming**: `*_test.go`.
- **Package**: Same package as the code being tested (e.g., `package ai`).
- **Framework**: Standard `testing` package.
- **Mocking**: Use `httptest.Server` for mocking API endpoints.
- **Coverage**: Aim for high coverage on logic-heavy packages like `ai` and `helper`.

## 3. Git Operations for Agents

- **No Execution**: Do NOT execute `git commit`, `git push`, or `git branch` commands unless explicitly instructed.
- **Worktrees**: If using the `brainstorming` or `writing-plans` skills, follow the instructions for creating git worktrees if applicable, but prefer working in the current directory if instructed.

## 4. Documentation

- Update `README.md` when adding new features or configuration options.
- Use `//` for single-line comments.
- Exported functions and types MUST have GoDoc-style comments.

  ```go
  // GenerateCommitMessage generates a commit message based on the provided diff.
  func (p *Provider) GenerateCommitMessage(...)
  ```

## 5. AI-Specific Guidelines

- **Context Management**: Use `discard` and `extract` tools to manage context window. Prune output that is no longer needed.
- **Verification**: Always verify your work using the `verification-before-completion` skill principles. Run tests and builds before claiming success.
- **Planning**: Use `brainstorming` for design and `writing-plans` for implementation planning on complex tasks.
