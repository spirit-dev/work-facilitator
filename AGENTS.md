# Agent Guidelines for work-facilitator

## Build, Lint, Test Commands

- **Build**: `make build-stand-alone` (outputs to `dist/work-facilitator`)
- **Run dev**: `make run-dev CMD_OPT="<args>"` or `cd src/work-facilitator && go run . <args>`
- **Test all**: `make test-dev` or `cd src/work-facilitator && go test -v ./...`
- **Test single**: `cd src/work-facilitator && go test -v -run TestName ./path/to/package`
- **Lint**: Pre-commit hooks run hadolint (Docker), markdownlint, gitleaks, detect-secrets
- **Install deps**: `make packages` or `cd src/work-facilitator && go mod tidy`

## Code Style & Conventions

- **Module**: `spirit-dev/work-facilitator` (Go 1.23+)
- **Imports**: Standard lib first, then external packages, then internal (`spirit-dev/work-facilitator/...`). Use aliases for common packages (e.g., `log "github.com/sirupsen/logrus"`, `c "spirit-dev/work-facilitator/work-facilitator/common"`)
- **Formatting**: Standard `gofmt` formatting (enforced by pre-commit)
- **Naming**: CamelCase for exported, camelCase for unexported. Package names lowercase single-word
- **Error handling**: Use `log.Fatalln()` for fatal errors, `log.Warningln()` for warnings, `log.Debugln()` for debug. Check errors explicitly
- **Comments**: Copyright header on all files. Use `//` for single-line, godoc-style for exported functions
- **Structure**: Commands in `cmd/`, helpers in `helper/`, common types in `common/`, integrations in `ticketing/`, AI providers in `ai/`
- **CLI**: Uses Cobra framework. Commands have PreRun for setup, Run for execution
- **Config**: Viper-based config from `~/.workflow.yaml`. Access via `helper.NewConfig()`
- **Tests**: Follow Go conventions with `_test.go` suffix. Run with `go test -v ./...`
- **AI Integration**: AI providers implement the `ai.Provider` interface. Supports OpenAI and Claude

## Pre-commit Hooks

Enforced: JSON/YAML validation, trailing whitespace, EOF fixer, no commits to main/master, Dockerfile linting, secrets detection, markdown linting. Use `-s` flag to skip if necessary.

## Tooling Preferences

- **Serena MCP**: When available, prefer Serena MCP tools for enhanced code navigation, symbol manipulation, and file operations over built-in alternatives

## Git Operations

- **No Execution**: Do NOT execute git commit or git branch operations. Only suggest the commands to be run.
