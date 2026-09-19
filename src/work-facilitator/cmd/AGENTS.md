# AGENTS.md — cmd

## Purpose

Cobra CLI commands for work-facilitator: aiCommit, commit, init, list, open, pause, status, use, completion, end, version.

## Ownership

- This package owns all user-facing CLI command definitions and flags
- Business logic that outlives a command belongs in `helper/`, `ai/`, or `ticketing/`

## Local Contracts

- Cobra pattern: commands define `Use`, `Short`, `Long`; setup goes in `PreRun`, execution in `Run`
- Config access via `helper.NewConfig()` (Viper, `~/.workflow.yaml`)
- Error handling: `log.Fatalln()` fatal, `log.Warningln()` warning, `log.Debugln()` debug
- Copyright header on all files

## Work Guidance

- New commands go in one file per command, named after the command
- Keep flags and UX consistent with existing commands (e.g. `initLazy.go` variants)

## Verification

- `cd src/work-facilitator && go test -v -run TestName ./work-facilitator/cmd` (existing tests: aiCommit, commit)

## Child DOX Index

None. All command files live flat in this directory.
