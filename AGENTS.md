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
- **GoDoc**: Exported symbols MUST have GoDoc-style comments; update `README.md` when adding new features or configuration options
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

---

## DOX framework

- DOX is highly performant AGENTS.md hierarchy installed here
- Agent must follow DOX instructions across any edits

### Core Contract

- AGENTS.md files are binding work contracts for their subtrees
- Work products, source materials, instructions, records, assets, and durable docs must stay understandable from the nearest applicable AGENTS.md plus every parent AGENTS.md above it

### Read Before Editing

1. Read the root AGENTS.md
2. Identify every file or folder you expect to touch
3. Walk from the repository root to each target path
4. Read every AGENTS.md found along each route
5. If a parent AGENTS.md lists a child AGENTS.md whose scope contains the path, read that child and continue from there
6. Use the nearest AGENTS.md as the local contract and parent docs for repo-wide rules
7. If docs conflict, the closer doc controls local work details, but no child doc may weaken DOX

Do not rely on memory. Re-read the applicable DOX chain in the current session before editing.

### Update After Editing

Every meaningful change requires a DOX pass before the task is done.

Update the closest owning AGENTS.md when a change affects:

- purpose, scope, ownership, or responsibilities
- durable structure, contracts, workflows, or operating rules
- required inputs, outputs, permissions, constraints, side effects, or artifacts
- user preferences about behavior, communication, process, organization, or quality
- AGENTS.md creation, deletion, move, rename, or index contents

Update parent docs when parent-level structure, ownership, workflow, or child index changes.
Update child docs when parent changes alter local rules. Remove stale or contradictory text immediately.
Small edits that do not change behavior or contracts may leave docs unchanged, but the DOX pass still must happen.

### Hierarchy

- Root AGENTS.md is the DOX rail: project-wide instructions, global preferences, durable workflow rules, and the top-level Child DOX Index
- Child AGENTS.md files own domain-specific instructions and their own Child DOX Index
- Each parent explains what its direct children cover and what stays owned by the parent
- The closer a doc is to the work, the more specific and practical it must be

### Child Doc Shape

- Create a child AGENTS.md when a folder becomes a durable boundary with its own purpose, rules, responsibilities, workflow, materials, or quality standards
- Work Guidance must reflect the current standards of the project or user instructions; if there are no specific standards or instructions yet, leave it empty
- Verification must reflect an existing check; if no verification framework exists yet, leave it empty and update it when one exists

Default section order:

- Purpose
- Ownership
- Local Contracts
- Work Guidance
- Verification
- Child DOX Index

### Style

- Keep docs concise, current, and operational
- Document stable contracts, not diary entries
- Put broad rules in parent docs and concrete details in child docs
- Prefer direct bullets with explicit names
- Do not duplicate rules across many files unless each scope needs a local version
- Delete stale notes instead of explaining history
- Trim obvious statements, repeated rules, misplaced detail, and warnings for risks that no longer exist

### Closeout

1. Re-check changed paths against the DOX chain
2. Update nearest owning docs and any affected parents or children
3. Refresh every affected Child DOX Index
4. Remove stale or contradictory text
5. Run existing verification when relevant
6. Report any docs intentionally left unchanged and why

### User Preferences

When the user requests a durable behavior change, record it here or in the relevant child AGENTS.md

### Child DOX Index

- `src/work-facilitator/cmd/AGENTS.md` — Cobra CLI commands (aiCommit, commit, init, list, open, pause, status, use, completion, end, version)
- `src/work-facilitator/ai/AGENTS.md` — AI provider implementations (OpenAI, Claude, Vertex AI, LlamaCPP) behind the `ai.Provider` interface
- `src/work-facilitator/helper/AGENTS.md` — Shared helpers: Viper config, display, git hooks, repo/ignore handling
- `src/work-facilitator/ticketing/AGENTS.md` — Ticketing integrations (GitLab, Jira)
- Not indexed separately: `src/work-facilitator/common/` (single-file shared types, covered by root Project Structure rules), `src/work-facilitator/main.go` (trivial entry point)
