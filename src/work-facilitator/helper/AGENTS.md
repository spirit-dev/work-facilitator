# AGENTS.md — helper

## Purpose

Shared helpers: Viper config (`config.go`), display, git hooks (`hooks.go`), repo management (`repo.go`), gitignore handling (`ignore.go`).

## Ownership

- This package owns cross-cutting utilities used by `cmd/`, `ai/`, `ticketing/`
- Config contract: Viper-based, file at `~/.workflow.yaml`, entry point `helper.NewConfig()`

## Local Contracts

- Error handling: `log.Fatalln()` fatal, `log.Warningln()` warning, `log.Debugln()` debug
- Copyright header on all files; gofmt formatting

## Work Guidance

- Keep helpers provider- and command-agnostic; anything specific to a command belongs in `cmd/`

## Verification

- `cd src/work-facilitator && go test -v ./work-facilitator/helper` (existing tests: hooks, ignore)

## Child DOX Index

None. Helper files live flat in this directory.
