# Task Completion Checklist for work-facilitator

When completing a development task, follow these steps:

## 1. Code Quality Checks

### Format Code

Code is automatically formatted by `gofmt` (enforced by pre-commit), but you can manually run:

```bash
cd src/work-facilitator && gofmt -w .
```

### Run Pre-commit Hooks

```bash
pre-commit run --all-files
```

This will check:

- JSON/YAML syntax
- Trailing whitespace
- End-of-file fixers
- No commits to main/master branches
- Dockerfile linting (hadolint)
- Secret detection (detect-secrets, gitleaks)
- Markdown linting

## 2. Build Verification

```bash
make build-stand-alone
```

Ensure the build completes without errors. Binary will be in `dist/work-facilitator`.

## 3. Testing

### Run All Tests

```bash
make test-dev
# or
cd src/work-facilitator && go test -v ./...
```

**Note**: Currently the project has no tests. If you added tests, ensure they all pass.

### Manual Testing

If you modified commands, test them manually:

```bash
make run-dev CMD_OPT="<command> <args>"
# or
cd src/work-facilitator && go run . <command> <args>
```

## 4. Dependencies

If you added new dependencies:

```bash
make packages
# or
cd src/work-facilitator && go mod tidy
```

## 5. Documentation

- Update README.md if you added new features or commands
- Add godoc-style comments for exported functions
- Ensure copyright headers are present in new files

## 6. Git Workflow

Before committing:

1. Ensure you're on a feature branch (not main/master)
2. Pre-commit hooks will run automatically on `git commit`
3. Follow commit message convention: `type: message`
   - Types: feat, fix, docs, style, refactor, test, build, chore, perf

## 7. Optional: Docker Build

If Docker-related changes were made:

```bash
make docker-lint
make docker-build
```

## Summary Checklist

- [ ] Code formatted (gofmt)
- [ ] Pre-commit hooks pass
- [ ] Build succeeds (`make build-stand-alone`)
- [ ] Tests pass (if any exist)
- [ ] Dependencies updated (`go mod tidy`)
- [ ] Documentation updated
- [ ] Commit message follows convention
- [ ] On feature branch (not main/master)
