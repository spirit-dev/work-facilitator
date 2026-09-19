# AGENTS.md — ticketing

## Purpose

Ticketing platform integrations: GitLab (`gitlab.go`), Jira (`jira.go`).

## Ownership

- This package owns all external ticketing API communication
- One file per platform; each platform's client/operations stay self-contained

## Local Contracts

- Credentials/config come from `helper.NewConfig()` (Viper, `~/.workflow.yaml`), never hardcoded
- Copyright header on all files; gofmt formatting

## Work Guidance

- New platform (e.g. GitHub issues): add one file, keep the public surface consistent with existing platforms so `cmd/` can treat them uniformly

## Verification

- No package-local tests exist yet; verification is `go build ./...` and downstream command tests until a test file is added

## Child DOX Index

None. Integration files live flat in this directory.
