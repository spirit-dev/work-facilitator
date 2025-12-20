# Implementation Plan: Ignore Patterns in Commit Command

## Objective

Ensure that files matching specific regex patterns (`out.y[a?]ml` and `out[0-9]+.y[a?]ml`) are always excluded from git commits, regardless of whether the `-a` (all files) flag is used.

## Requirements Summary

- **Scope**: Ignore files only during commit operation (programmatically, not via .gitignore)
- **Pattern Type**: Regex patterns
- **Configuration**: Default patterns in code + ability to override via `~/.workflow.yaml`
- **Behavior**: Exclude from add operations AND unstage if already staged
- **Error Handling**: Fail the commit if ignored files are staged

---

## Architecture Overview

### Default Patterns (Hardcoded)

```go
// In common/common.go or new file helper/ignore.go
const (
    DefaultIgnorePattern1 = `out\.ya?ml$`      // Matches: out.yaml, out.yml
    DefaultIgnorePattern2 = `out\d+\.ya?ml$`   // Matches: out123.yaml, out456.yml
)
```

### Configuration Structure

Users can override defaults in `~/.workflow.yaml`:

```yaml
global:
  commit_ignore_patterns:
    - "out\\.ya?ml$"
    - "out\\d+\\.ya?ml$"
    - "custom_pattern.*\\.tmp$" # Optional: additional patterns
```

---

## Implementation Steps

### Phase 1: Configuration & Data Structures

#### Step 1.1: Update `common/common.go`

**File**: `src/work-facilitator/work-facilitator/common/common.go`

Add to `Config` struct:

```go
type Config struct {
    // ... existing fields ...
    CommitIgnorePatterns []string  // Regex patterns to ignore during commit
}
```

Add constants for default patterns:

```go
const (
    DefaultCommitIgnorePattern1 = `out\.ya?ml$`
    DefaultCommitIgnorePattern2 = `out\d+\.ya?ml$`
)
```

#### Step 1.2: Update `helper/config.go`

**File**: `src/work-facilitator/helper/config.go`

In `NewConfig()` function:

1. Read `global.commit_ignore_patterns` from viper (optional)
2. If not provided, use default patterns
3. Validate all patterns are valid regex
4. Store in Config struct

```go
// Pseudo-code
commitIgnorePatterns := viper.GetStringSlice("global.commit_ignore_patterns")
if len(commitIgnorePatterns) == 0 {
    commitIgnorePatterns = []string{
        c.DefaultCommitIgnorePattern1,
        c.DefaultCommitIgnorePattern2,
    }
}
// Validate patterns are valid regex
for _, pattern := range commitIgnorePatterns {
    if _, err := regexp.Compile(pattern); err != nil {
        log.Fatalln("Invalid regex pattern in commit_ignore_patterns: " + pattern)
    }
}
```

---

### Phase 2: Core Ignore Logic

#### Step 2.1: Create `helper/ignore.go` (New File)

**File**: `src/work-facilitator/helper/ignore.go`

This file will contain all ignore-pattern-related logic:

**Function 1: `FileMatchesIgnorePattern()`**

```go
// FileMatchesIgnorePattern checks if a file path matches any ignore pattern
func FileMatchesIgnorePattern(filePath string, patterns []string) bool
```

- Takes file path and list of regex patterns
- Returns true if file matches ANY pattern
- Handles regex compilation errors gracefully

**Function 2: `FilterIgnoredFiles()`**

```go
// FilterIgnoredFiles returns only files that DON'T match ignore patterns
func FilterIgnoredFiles(files []string, patterns []string) []string
```

- Takes list of file paths and patterns
- Returns filtered list (excluding matched files)
- Used by RepoAddAllFiles()

**Function 3: `GetStagedIgnoredFiles()`**

```go
// GetStagedIgnoredFiles returns staged files that match ignore patterns
func GetStagedIgnoredFiles(status git.Status, patterns []string) []string
```

- Takes git.Status and patterns
- Returns list of staged files matching patterns
- Used for validation before commit

**Function 4: `UnstageIgnoredFiles()`**

```go
// UnstageIgnoredFiles removes ignored files from staging area
func UnstageIgnoredFiles(patterns []string) error
```

- Gets current git status
- Identifies staged files matching patterns
- Uses go-git to unstage them
- Returns error if unstaging fails

---

### Phase 3: Modify Git Operations

#### Step 3.1: Update `helper/repo.go` - `RepoAddAllFiles()`

**Current Code**:

```go
func RepoAddAllFiles() {
    w, err := repo.Worktree()
    // ... error handling ...
    w.AddWithOptions(&git.AddOptions{All: true})
}
```

**Modified Code**:

1. Get current git status
2. Filter out files matching ignore patterns
3. Add only non-ignored files individually (since go-git's AddWithOptions doesn't support exclusions)

**Approach**:

- Use `RepoStatus()` to get all modified/untracked files
- Filter using `FilterIgnoredFiles()`
- Add each file individually using `w.Add(filePath)`

#### Step 3.2: Update `helper/repo.go` - `RepoCommit()`

**Current Code**:

```go
func RepoCommit(message string) {
    w, err := repo.Worktree()
    // ... error handling ...
    w.Commit(message, &git.CommitOptions{AllowEmptyCommits: false})
}
```

**Modified Code**:

1. Before committing, check for staged ignored files
2. If found, unstage them
3. If unstaging fails, log error and fail commit
4. Proceed with commit

**Implementation**:

```go
func RepoCommit(message string, ignorePatterns []string) {
    // ... existing code ...

    // Check for and unstage ignored files
    stagedIgnored := GetStagedIgnoredFiles(RepoStatus(), ignorePatterns)
    if len(stagedIgnored) > 0 {
        log.Warningln("Found staged files matching ignore patterns:")
        for _, f := range stagedIgnored {
            log.Warningln("  - " + f)
        }

        err := UnstageIgnoredFiles(ignorePatterns)
        if err != nil {
            log.Fatalln("Failed to unstage ignored files: " + err.Error())
        }
        log.Fatalln("Commit aborted: ignored files were staged and have been unstaged")
    }

    // ... rest of commit logic ...
}
```

---

### Phase 4: Update Command Layer

#### Step 4.1: Update `cmd/commit.go`

**Changes**:

1. Pass `RootConfig.CommitIgnorePatterns` to `RepoAddAllFiles()` and `RepoCommit()`
2. Update function signatures to accept patterns

**Modified Functions**:

```go
// In commitCommand()
if allFilesCommitArg {
    helper.SpinUpdateDisplay("Git add all files")
    helper.RepoAddAllFiles(RootConfig.CommitIgnorePatterns)
}

helper.SpinUpdateDisplay("Git commit")
helper.RepoCommit(commitMessageCommit, RootConfig.CommitIgnorePatterns)
```

---

### Phase 5: Testing Strategy

#### Step 5.1: Unit Tests for `helper/ignore.go`

**File**: `src/work-facilitator/helper/ignore_test.go`

Test cases:

1. `TestFileMatchesIgnorePattern_MatchesPattern()` - File matches pattern
2. `TestFileMatchesIgnorePattern_NoMatch()` - File doesn't match
3. `TestFileMatchesIgnorePattern_MultiplePatterns()` - Multiple patterns
4. `TestFilterIgnoredFiles_FiltersCorrectly()` - Filtering works
5. `TestFilterIgnoredFiles_EmptyList()` - Empty input
6. `TestGetStagedIgnoredFiles_FindsStagedIgnored()` - Finds staged ignored files
7. `TestUnstageIgnoredFiles_UnstagesSuccessfully()` - Unstaging works

#### Step 5.2: Integration Tests

Test the full commit workflow with ignored files

---

## Files to Modify/Create

| File                                                     | Action | Changes                                                                             |
| -------------------------------------------------------- | ------ | ----------------------------------------------------------------------------------- |
| `src/work-facilitator/work-facilitator/common/common.go` | Modify | Add `CommitIgnorePatterns []string` to Config struct; Add default pattern constants |
| `src/work-facilitator/helper/config.go`                  | Modify | Load patterns from config; Validate regex; Set defaults                             |
| `src/work-facilitator/helper/ignore.go`                  | Create | New file with all ignore pattern logic                                              |
| `src/work-facilitator/helper/repo.go`                    | Modify | Update `RepoAddAllFiles()` and `RepoCommit()` signatures and logic                  |
| `src/work-facilitator/cmd/commit.go`                     | Modify | Pass patterns to helper functions                                                   |
| `src/work-facilitator/helper/ignore_test.go`             | Create | Unit tests for ignore logic                                                         |

---

## Configuration Example

**~/.workflow.yaml** (optional override):

```yaml
global:
  app_name: "work-facilitator"
  script_name: "wf"
  version: "1.0.0"
  # ... other config ...

  # Optional: override default ignore patterns
  commit_ignore_patterns:
    - "out\\.ya?ml$"
    - "out\\d+\\.ya?ml$"
    - "temp.*\\.log$" # Additional custom pattern
```

If not specified, defaults are used:

- `out\.ya?ml$` (matches: out.yaml, out.yml)
- `out\d+\.ya?ml$` (matches: out123.yaml, out456.yml, etc.)

---

## Error Handling & Logging

### Scenarios

1. **Invalid regex pattern in config**: Fatal error during config load
2. **Ignored files in staging area**: Warning logged, files unstaged, commit fails
3. **Unstaging fails**: Fatal error
4. **No ignored files**: Normal flow continues

### Log Messages

```txt
WARN: Found staged files matching ignore patterns:
      - out.yaml
      - out123.yml
FATAL: Commit aborted: ignored files were staged and have been unstaged
```

---

## Backward Compatibility

- **No breaking changes**: Existing code continues to work
- **New optional config**: Users don't need to configure anything; defaults work
- **Function signatures**: New parameters added to existing functions (requires updates in callers)

---

## Implementation Order

1. **Step 1**: Update `common/common.go` (add struct field + constants)
2. **Step 2**: Create `helper/ignore.go` (core logic)
3. **Step 3**: Update `helper/config.go` (load patterns)
4. **Step 4**: Update `helper/repo.go` (integrate ignore logic)
5. **Step 5**: Update `cmd/commit.go` (pass patterns)
6. **Step 6**: Create tests
7. **Step 7**: Update documentation

---

## Potential Challenges & Mitigations

| Challenge                                           | Mitigation                               |
| --------------------------------------------------- | ---------------------------------------- |
| go-git doesn't support exclusions in AddWithOptions | Add files individually after filtering   |
| Performance with many files                         | Cache compiled regex patterns            |
| Complex regex patterns                              | Validate at config load time             |
| Existing staged files                               | Unstage before commit with clear warning |

---

## Success Criteria

✅ Files matching `out\.ya?ml$` are excluded from commits
✅ Files matching `out\d+\.ya?ml$` are excluded from commits
✅ Works with `-a` flag (all files)
✅ Works without `-a` flag (manual staging)
✅ Unstages ignored files if already staged
✅ Fails commit with clear error message if ignored files are staged
✅ Patterns configurable via `~/.workflow.yaml`
✅ Default patterns work without config
✅ All tests pass
✅ No breaking changes to existing functionality
