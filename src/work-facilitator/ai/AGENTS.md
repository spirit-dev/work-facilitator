# AGENTS.md — ai

## Purpose

AI provider implementations for work-facilitator: OpenAI, Claude, Vertex AI, LlamaCPP.

## Ownership

- This package owns all AI provider communication, prompt building (`prompt.go`), and retry logic (`retry.go`)
- The `ai.Provider` interface (in `types.go`) is the contract every provider must satisfy

## Local Contracts

- New providers implement `ai.Provider` and live in one file per provider (see `llamacpp.go` for the latest example)
- Providers must expose `Validate()` checking required configuration; use the shared `buildPrompt` helper for prompt consistency
- Handle timeouts and context cancellation properly
- Copyright header on all files; gofmt formatting; aliased imports per root guidelines

## Work Guidance

- Adding a provider: mirror the structure of `openai.go`/`claude.go`; no provider-specific logic may leak into `cmd/` or `helper/`

## Verification

- `cd src/work-facilitator && go test -v ./work-facilitator/ai` (existing tests: types, vertexai)

## Child DOX Index

None. Provider files live flat in this directory.
