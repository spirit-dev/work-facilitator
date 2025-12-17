# Commit Ignore Patterns - Quick Summary

## What We're Building

A feature that prevents specific files from being committed, regardless of how they're staged.

## Default Patterns

- `out\.ya?ml$` → matches `out.yaml`, `out.yml`
- `out\d+\.ya?ml$` → matches `out123.yaml`, `out456.yml`, etc.

## User Workflow (After Implementation)

### Scenario 1: User runs `commit -a "message"`

```txt
1. RepoAddAllFiles() is called
2. System filters out files matching ignore patterns
3. Only non-ignored files are added to staging
4. Commit proceeds normally
```

### Scenario 2: User manually stages ignored file, then runs `commit "message"`

```txt
1. RepoCommit() is called
2. System detects staged files matching ignore patterns
3. System unstages those files
4. System logs warning and FAILS the commit
5. User must fix and retry
```

### Scenario 3: User configures custom patterns in ~/.workflow.yaml

```yaml
global:
  commit_ignore_patterns:
    - "out\\.ya?ml$"
    - "out\\d+\\.ya?ml$"
    - "custom_pattern.*"
```

- Custom patterns override defaults
- All patterns are validated at startup

## Code Changes Overview

### New Files

- `helper/ignore.go` - Core ignore pattern logic
- `helper/ignore_test.go` - Unit tests

### Modified Files

- `common/common.go` - Add CommitIgnorePatterns field
- `helper/config.go` - Load patterns from config
- `helper/repo.go` - Integrate ignore logic into RepoAddAllFiles() and RepoCommit()
- `cmd/commit.go` - Pass patterns to helper functions

## Key Functions to Implement

| Function                     | Purpose                                |
| ---------------------------- | -------------------------------------- |
| `FileMatchesIgnorePattern()` | Check if file matches any pattern      |
| `FilterIgnoredFiles()`       | Remove ignored files from list         |
| `GetStagedIgnoredFiles()`    | Find staged files matching patterns    |
| `UnstageIgnoredFiles()`      | Remove ignored files from staging area |

## Integration Points

```txt
commit.go
  ├─ RepoAddAllFiles(patterns)
  │   └─ FilterIgnoredFiles() → only add non-ignored files
  │
  └─ RepoCommit(message, patterns)
      ├─ GetStagedIgnoredFiles() → check for violations
      ├─ UnstageIgnoredFiles() → remove violations
      └─ Fail if violations found
```

## Error Handling

| Situation                | Action                             |
| ------------------------ | ---------------------------------- |
| Invalid regex in config  | Fatal error at startup             |
| Ignored files in staging | Unstage + fail commit with warning |
| Unstaging fails          | Fatal error                        |

## Testing Strategy

- Unit tests for each ignore function
- Integration tests for full commit workflow
- Test with both default and custom patterns
- Test edge cases (empty files, special characters, etc.)

## Configuration (Optional)

Users can override defaults in `~/.workflow.yaml`:

```yaml
global:
  commit_ignore_patterns:
    - "pattern1"
    - "pattern2"
```

If not specified, hardcoded defaults are used.

## Success Metrics

✅ Ignored files never make it into commits
✅ Works with `-a` flag
✅ Works without `-a` flag
✅ Clear error messages when violations occur
✅ Configurable patterns
✅ No breaking changes
✅ Full test coverage
