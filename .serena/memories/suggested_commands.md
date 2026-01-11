# Suggested Commands for work-facilitator

## Development Commands

### Build

```bash
make build-stand-alone          # Build binary to dist/work-facilitator
make clean-sdtl                 # Clean build artifacts
```

### Run

```bash
make run-dev CMD_OPT="<args>"   # Run in development mode with arguments
cd src/work-facilitator && go run . <args>  # Direct go run
make run-stand-alone            # Build and run standalone binary
```

### Testing

```bash
make test-dev                   # Run all tests
cd src/work-facilitator && go test -v ./...  # Run all tests with verbose output
cd src/work-facilitator && go test -v -run TestName ./path/to/package  # Run single test
```

### Dependencies

```bash
make packages                   # Install/update Go modules
cd src/work-facilitator && go mod tidy  # Tidy Go modules
```

### Installation

```bash
make install                    # Build and install to /usr/local/bin (requires sudo)
```

### Completion

```bash
make completion                 # Generate shell completion (zsh)
```

## Quality Assurance

### Pre-commit Hooks

```bash
pre-commit run --all-files      # Run all pre-commit hooks
pre-commit install              # Install pre-commit hooks
```

Pre-commit hooks enforce:

- JSON/YAML validation
- Trailing whitespace removal
- End-of-file fixer
- No commits to main/master
- Dockerfile linting (hadolint)
- Secret detection (detect-secrets, gitleaks)
- Markdown linting (markdownlint)

### Docker

```bash
make docker-lint                # Lint Dockerfile
make docker-build               # Build Docker images (amd64, arm64)
make docker-run                 # Run container interactively
```

## System Commands (Linux)

Standard Linux commands available:

- `git` - Version control
- `ls`, `cd`, `pwd` - File navigation
- `grep`, `find` - Search
- `cat`, `less` - File viewing
- `make` - Build automation
