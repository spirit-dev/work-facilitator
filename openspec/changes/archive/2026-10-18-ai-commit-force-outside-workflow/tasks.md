# Tasks

## 1. Branch Resolution in ai-commit

- [x] 1.1 In `src/work-facilitator/cmd/aiCommit.go`, compute `effectiveBranch` and `effectivePreMessage` at the start of `aiCommitCommand`: use `RootRepo.CurrentWorkflowData` when `HasCurrentWorkflow`, otherwise derive the branch from `helper.RepoHead()` and use an empty pre-message; verify with `go build ./...`
- [x] 1.2 Replace direct reads of `RootRepo.CurrentWorkflowData.Branch` (AI `options.BranchName`, `TestStandard` branch argument) and `.Commit` (pre-message) with the effective values; verify existing `ai-commit` tests still pass via `go test ./cmd/`
- [x] 1.3 Pass `effectiveBranch` to `helper.RepoPush` for both workflow and non-workflow paths; verify by reading the call site matches the design decision

## 2. Tests

- [x] 2.1 Add a test in `src/work-facilitator/cmd/aiCommit_test.go` covering effective branch resolution outside a workflow (no workflow data, branch derived from HEAD, empty pre-message); verify with `go test -v -run TestAiCommit ./cmd/`
- [x] 2.2 Add a test covering the workflow-active path keeps using `CurrentWorkflowData.Branch` and its commit prefix (regression guard); verify with `go test ./cmd/`

## 3. Validation and Docs

- [x] 3.1 Run full test suite `make test-dev` (or `cd src/work-facilitator && go test ./...`) and confirm it passes
- [x] 3.2 Update README.md if the `-f` flag behavior for `ai-commit` is documented, clarifying it commits/pushes on the current branch outside a workflow; verify with `markdownlint README.md` if hooks run it
