# Design: AI-Assisted Commit — Diff Generation

## Problem

`GetStagedDiff()` in `helper/repo.go` has a bug: it reads file content from `w.Filesystem.Open(filePath)` (the working directory), not from the git index (staged blobs). This means unstaged modifications leak into the AI context, producing commit messages that describe changes not actually being committed.

Additionally, the current diff format is a naive line-by-line comparison without context lines or hunk headers (`@@` markers), which reduces AI message quality.

## Solution

### Core Fix: Read from Index Blobs

```
BEFORE (bug):                        AFTER (fix):
───────────────────────────          ───────────────────────────
idx.Entries → entry.Name             idx.Entries → entry.Name
  │                                    │
  └─ w.Filesystem.Open(filePath)       └─ repo.Storer.EncodedObject(
       ↑ reads WORKING DIRECTORY              plumbing.BlobObject,
                                              entry.Hash)
                                            ↑ reads STAGED BLOB
```

The git index entry has an `entry.Hash` field pointing to the staged blob in the object store. The fix reads from `repo.Storer.EncodedObject()` instead of `w.Filesystem`.

**go-git imports needed** (no new dependencies — `plumbing` already imported):
```go
import "github.com/go-git/go-git/v5/plumbing"
```

### New Flag: `-U` / `--include-unstaged`

For developers who want the AI to see the full current state of staged files (including unstaged modifications on top), the `-U` flag switches from index-blob reading to working-tree reading — but only for files that are in the index.

```
                          Without -U           With -U
                          ─────────────        ─────────
Staged file content:      index blob           working tree file
Unstaged file content:    excluded             excluded
Untracked file content:   excluded             excluded
```

A file must have a staged index entry to be included at all. The flag controls *which version* of that file is read.

### Diff Format Upgrade

Replace the naive line-by-line comparison:

```
BEFORE (naive):                      AFTER (unified diff):
────────────────────                 ──────────────────────────
-old line                            @@ -10,7 +10,8 @@
+new line                              context line
                                       context line
                                      -old line
                                      +new line
                                       context line
```

The new format includes:
- `@@ -start,count +start,count @@` hunk headers
- 3 lines of context before and after changes
- Proper `--- a/file` / `+++ b/file` headers (already present)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      cmd/aiCommit.go                        │
│                                                             │
│  if allFilesAICommitArg:                                    │
│    helper.RepoAddAllFiles(ignorePatterns)                   │
│                                                             │
│  if includeUnstagedAICommitArg:                             │
│    diff = helper.GetWorkingTreeDiff()   ← reads work-tree   │
│  else:                                                      │
│    diff = helper.GetStagedDiff()        ← reads index blobs │
│                                                             │
│  diff = helper.FilterDiffByPatterns(diff, excludePatterns)  │
│  message = aiProvider.GenerateCommitMessage(ctx, diff, ...) │
│  helper.RepoCommit(message, ignorePatterns)                 │
└─────────────────────────────────────────────────────────────┘
```

### Functions in `helper/repo.go`

| Function | Behavior | Used when |
|----------|----------|-----------|
| `GetStagedDiff()` | HEAD tree vs index blobs (proper unified diff) | Default (no flag) |
| `GetWorkingTreeDiff()` | HEAD tree vs working directory (proper unified diff) | `-U` flag |

Both functions only consider files present in the git index.

## Affected Files

| File | Change |
|------|--------|
| `helper/repo.go` | Fix `GetStagedDiff()`: read from index blobs; implement proper unified diff; add `GetWorkingTreeDiff()` |
| `cmd/aiCommit.go` | Add `includeUnstagedAICommitArg` variable, `-U` flag definition, conditional diff selection |
