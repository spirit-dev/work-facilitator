# Spec Delta

## Purpose

Defines the behavior of AI-assisted commits performed with the force flag (`ai-commit -f`) when the repository has no active workflow, so users can generate and push AI-written commit messages on any branch.

## ADDED Requirements

### Requirement: Forced AI-Assisted Commit Outside a Workflow

**User Story:** As a developer, I want to run `work-facilitator ai-commit -f` on a branch with no active workflow, so that I can commit with an AI-generated message without starting a workflow.

The system SHALL complete the full commit-and-push flow when `-f` is used outside an active workflow, operating on the current git branch.

#### Acceptance Criteria

1. WHEN user runs `work-facilitator ai-commit -f` and no current workflow exists THEN system SHALL complete the commit and push flow instead of exiting
2. WHEN no current workflow exists THEN system SHALL resolve the current git branch and use it as the branch context for AI message generation
3. WHEN no current workflow exists and the commit standard is enforced THEN system SHALL validate the message against the standard using the current git branch as branch context and without any workflow commit prefix
4. WHEN no current workflow exists and push is requested THEN system SHALL push the current git branch to the remote
5. WHEN no current workflow exists and the repository has no resolvable HEAD (e.g., empty repository) THEN system SHALL fail with a clear error instead of pushing an invalid refspec
6. WHEN a current workflow exists THEN system SHALL behave exactly as without `-f` (workflow branch and commit prefix used)

#### Scenario: Forced Commit on a Plain Branch

1. Developer is on branch `feature-x` with no active workflow
2. Developer stages changes and runs `work-facilitator ai-commit -f`
3. System generates an AI commit message using `feature-x` as branch context
4. Developer accepts the message
5. System commits and pushes `feature-x` to the remote
6. System displays success

#### Scenario: Forced Commit Without Push

1. Developer is on branch `feature-x` with no active workflow
2. Developer runs `work-facilitator ai-commit -f -n`
3. System generates, validates, and commits with the AI message
4. System does NOT push
5. System displays success

#### Scenario: Forced Commit With Enforced Standard

1. Developer has `enforce_standard: true` with a commit expression requiring a type prefix
2. Developer is on branch `main` with no active workflow
3. Developer runs `work-facilitator ai-commit -f`
4. System validates the AI message against the commit standard using `main` as branch context and no workflow commit prefix
5. A compliant message is committed and pushed to `main`

#### Scenario: Forced Commit Without Resolvable HEAD

1. Repository has no commits (no HEAD)
2. Developer runs `work-facilitator ai-commit -f`
3. System reports a clear error and exits without committing or pushing

#### Scenario: Active Workflow Takes Precedence

1. Developer runs `ai-commit` inside an active workflow
2. System uses the workflow branch and commit prefix as before
3. Behavior is unchanged relative to running without `-f`
