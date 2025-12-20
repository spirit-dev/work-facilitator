# work-facilitator Project Overview

## Purpose

work-facilitator is a CLI tool that manages Git workflow interactions between JIRA, GitHub, and GitLab. It enforces standards for branch naming and commit message conventions to ensure compliance with semantic release practices.

## Key Features

- **init**: Initialize work linked with JIRA or GitLab issues
- **commit**: Commit changes with properly prefixed messages
- **end**: Close a work/branch
- **list**: List created works
- **open**: Open browser to repository
- **pause**: Pause current work
- **status**: Show status of current work
- **use**: Resume a paused work
- **initLazy**: Create work based on JIRA/GitLab information
- **completion**: Generate shell completion for Linux/Mac

## Tech Stack

- **Language**: Go 1.23+ (module: `spirit-dev/work-facilitator`)
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **Configuration**: Viper (github.com/spf13/viper) - reads from `~/.workflow.yaml`
- **Logging**: Logrus (github.com/sirupsen/logrus)
- **Git Operations**: go-git (github.com/go-git/go-git/v5)
- **Integrations**:
  - JIRA: github.com/andygrunwald/go-jira
  - GitLab: github.com/xanzy/go-gitlab
- **UI**: pterm for terminal output

## Project Structure

```txt
src/work-facilitator/
├── cmd/          # Cobra commands (init, commit, end, list, open, pause, status, use, etc.)
├── helper/       # Helper functions (config, display, repo operations)
├── common/       # Common types and constants
├── ticketing/    # JIRA and GitLab integrations
└── main.go       # Entry point
```

## Configuration

- Config file: `~/.workflow.yaml`
- Supports both JIRA and GitLab ticketing systems (mutually exclusive)
- Enforces branch naming: `(feat|fix|release|renovate)/*`
- Enforces commit messages: `(feat|fix|docs|style|refactor|test|build|chore|perf): message`
- Type mapping for branch to commit type conversion
