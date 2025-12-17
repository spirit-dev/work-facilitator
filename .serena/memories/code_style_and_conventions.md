# Code Style and Conventions for work-facilitator

## Go Module

- Module name: `spirit-dev/work-facilitator`
- Go version: 1.23+ (toolchain: go1.23.10)

## Import Organization

Imports must be organized in three groups (separated by blank lines):

1. Standard library packages
2. External/third-party packages
3. Internal packages (`spirit-dev/work-facilitator/...`)

### Import Aliases

Use aliases for commonly used packages:

- `log "github.com/sirupsen/logrus"` - for logging
- `c "spirit-dev/work-facilitator/work-facilitator/common"` - for common package

Example:

```go
import (
    "fmt"
    "os"

    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"

    c "spirit-dev/work-facilitator/work-facilitator/common"
    "spirit-dev/work-facilitator/work-facilitator/helper"
)
```

## Formatting

- Use standard `gofmt` formatting (enforced by pre-commit hooks)
- No manual formatting needed - `gofmt` handles it

## Naming Conventions

- **Exported symbols**: CamelCase (e.g., `NewConfig`, `RootConfig`)
- **Unexported symbols**: camelCase (e.g., `initConfig`, `homeDir`)
- **Package names**: lowercase, single-word (e.g., `cmd`, `helper`, `common`, `ticketing`)
- **Constants**: CamelCase for exported, camelCase for unexported

## Error Handling

Use logrus for all logging and error handling:

- `log.Fatalln()` - Fatal errors that should terminate the program
- `log.Warningln()` - Warnings that don't stop execution
- `log.Debugln()` - Debug information
- `log.Infoln()` - Informational messages
- Always check errors explicitly (don't ignore them)

Example:

```go
if err != nil {
    log.Fatalln("fatal error config file: %w", err)
}
```

## Comments and Documentation

- **Copyright header**: All files must have copyright header at the top

  ```go
  /*
  Copyright © 2024 Jean Bordat bordat.jean@gmail.com
  */
  ```

- **Single-line comments**: Use `//` for single-line comments
- **Exported functions**: Use godoc-style comments (comment starts with function name)

  ```go
  // NewConfig creates and returns a new Config instance
  func NewConfig() c.Config {
  ```

## Project Structure

- `cmd/` - Cobra command implementations (one file per command)
- `helper/` - Helper functions (config, display, repo operations)
- `common/` - Common types, structs, and constants
- `ticketing/` - External integrations (JIRA, GitLab)
- `main.go` - Application entry point

## CLI Framework (Cobra)

- Commands use `PreRun` for setup/validation
- Commands use `Run` for main execution logic
- Command variables are package-level in their respective cmd files
- Use `cobra.Command` struct with `Use`, `Short`, `Long`, `Args`, `ValidArgs`

## Configuration (Viper)

- Config file: `~/.workflow.yaml`
- Access config via `helper.NewConfig()` which returns `common.Config` struct
- Use `viper.GetString()`, `viper.GetBool()`, etc. to read config values
- Environment variables prefixed with `WF_` (uppercase)

## Testing

- Currently no tests exist in the project
- When adding tests, follow Go conventions:
  - Test files: `*_test.go` suffix
  - Test functions: `func TestXxx(t *testing.T)`
  - Place tests in the same package as the code being tested
