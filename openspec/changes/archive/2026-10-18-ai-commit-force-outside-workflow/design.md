# Design

## Context

`aiCommitPreRunCommand` already passes when `-f` is set and no workflow exists, but `aiCommitCommand` reads `RootRepo.CurrentWorkflowData` (a zero-value struct when `HasCurrentWorkflow` is false) in three places: AI prompt branch context, standard validation, and push. `helper.RepoHead()` already exists and returns the current branch reference; `helper.RepoPush(pubKey, branch)` takes an explicit branch string. See proposal.md for motivation and the delta spec for required behavior.

## Goals / Non-Goals

**Goals:**
- One branch-resolution point in `aiCommitCommand` that yields the branch to use for AI context, validation, and push.
- Outside a workflow: use current git HEAD branch, no workflow commit prefix, push current branch.
- Inside a workflow: behavior identical to today.

**Non-Goals:**
- No changes to the plain `commit` command (it has the same latent issue but is out of scope; flag it separately if wanted).
- No new commands or aliases for `ai-commit`.
- No changes to AI providers or prompt building beyond the branch context value.

## Decisions

- **Resolve branch in `aiCommitCommand`, not `NewRepo`.** Compute an effective branch once at the top of the run: if `RootRepo.HasCurrentWorkflow`, use `RootRepo.CurrentWorkflowData.Branch`; otherwise derive it from `RepoHead().Name.Short()` when `-f` was used. Alternatives: populating `CurrentWorkflowData` in `NewRepo` (mutates shared repo state, affects other commands — rejected); resolving at each use site (three call sites, easy to miss one — rejected).
- **Effective pre-message and branch for validation.** Introduce local variables `effectiveBranch` and `effectivePreMessage` in `aiCommitCommand`, replacing direct `RootRepo.CurrentWorkflowData` reads. Outside a workflow: `effectivePreMessage = ""`. This keeps the existing `TestStandard` call unchanged in shape.
- **Fail fast on unresolvable HEAD.** If no workflow and `RepoHead()` errors (empty repo, detached HEAD edge cases), `log.Fatalln` with a clear message, matching existing error style in `helper/repo.go`. `RepoHead()` already fatals internally; the command relies on that and needs no extra check beyond using its result.
- **Push target = effective branch.** Pass `effectiveBranch` to `RepoPush` in both workflow and non-workflow cases; the workflow case is unchanged because `effectiveBranch == CurrentWorkflowData.Branch` there.

## Risks / Trade-offs

- [Detached HEAD yields a short hash, not a branch name] → Pushing `refs/heads/<hash>` fails loudly at push time; acceptable for a force flag. Document in flag help if needed.
- [`commit -f` shares the same latent bug] → Explicitly out of scope; noted for a possible follow-up change.

## Migration Plan

No migration. Pure behavior fix behind an existing flag; rollback is reverting the single file change.

## Open Questions

None.
