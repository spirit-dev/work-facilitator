# Plan: Add Uncommitted Files Detection to `pause` Command

## Overview

Add the same uncommitted files detection and force flag behavior from the `end` command to the `pause` command. This will ensure consistency across workflow-altering commands.

## Current State Analysis

### `pause` Command Structure

- **File**: `src/work-facilitator/cmd/pause.go`
- **Lines**: 95 total
- **Current Flags**:
  - `-m, --master`: Go to master branch instead of ref-branch
- **Phases**:
  - **PreRun** (`pausePreRunCommand`): Lines 32-63
    - Validates current workflow exists
    - Determines checkout target (master or ref-branch)
  - **Run** (`pauseCommand`): Lines 65-87
    - Performs git checkout
    - Performs git pull
    - Deletes current workflow config
    - Displays goodbye message

### Key Differences from `end` Command

1. `pause` doesn't delete the branch, only pauses it
2. `pause` requires an active workflow (no `-w` flag to specify workflow)
3. `pause` has a `-m` flag to choose between master or ref-branch
4. `pause` is simpler - no branch deletion, just checkout and config update

## Implementation Plan

### Phase 1: Add Force Flag Variable

**File**: `src/work-facilitator/cmd/pause.go`
**Location**: Lines 15-21 (var block)

**Changes**:

```go
var (
 // Cmd Args
 masterPauseArg string
 forcePause bool  // NEW: Add force flag variable

 // local
 checkoutToPause string
)
```

### Phase 2: Add Uncommitted Files Check in PreRun

**File**: `src/work-facilitator/cmd/pause.go`
**Location**: After line 59 (after checkoutToPause is determined)

**Changes**:

- Add check for uncommitted files (similar to `end` command)
- Respect `RootConfig.UncommittedFilesDetection` setting
- Skip check if `forcePause` flag is set
- Handle all 4 modes: disabled, warning, fatal, interactive
- Display warning if force mode is active

**Code Pattern** (from `end.go` lines 76-103):

```go
// Check for uncommitted files during validation phase
if RootConfig.UncommittedFilesDetection != "disabled" && !forcePause {
    log.Debug("Checking for uncommitted files...")
    uncommittedFiles, hasUncommitted, err := helper.RepoCheckUncommittedFiles()
    if err != nil {
        helper.SpinStopDisplay("fail")
        log.Fatalln("Error checking repository status:", err)
    }

    if hasUncommitted {
        helper.SpinStopDisplay("warning")
        helper.DisplayUncommittedFiles(uncommittedFiles)

        switch RootConfig.UncommittedFilesDetection {
        case "fatal":
            log.Fatalln("Uncommitted files detected. Please commit or stash changes before pausing workflow.")
        case "warning":
            helper.SpinSideNoteDisplay("Warning: Uncommitted files detected")
            helper.SpinStartDisplay("Verifications - pause...")
        case "interactive":
            if !helper.PromptUserConfirmation("Continue with uncommitted files?") {
                log.Fatalln("Operation cancelled by user")
            }
            helper.SpinStartDisplay("Verifications - pause...")
        }
    }
} else if forcePause && RootConfig.UncommittedFilesDetection != "disabled" {
    log.Debug("Force flag set, skipping uncommitted files check...")
    helper.SpinSideNoteDisplay("Warning: Skipping uncommitted files check (force mode)")
}
```

### Phase 3: Add Re-check in Run Phase (Optional but Recommended)

**File**: `src/work-facilitator/cmd/pause.go`
**Location**: After line 67 (after SpinStartDisplay)

**Changes**:

- Add re-check for fatal mode only (like in `end` command)
- Ensures files didn't change between PreRun and Run
- Skip if `forcePause` is set

**Code Pattern** (from `end.go` lines 115-129):

```go
// Re-check for uncommitted files in case files changed between PreRun and Run
if RootConfig.UncommittedFilesDetection == "fatal" && !forcePause {
    log.Debug("Re-checking for uncommitted files before git operations...")
    uncommittedFiles, hasUncommitted, err := helper.RepoCheckUncommittedFiles()
    if err != nil {
        helper.SpinStopDisplay("fail")
        log.Fatalln("Error checking repository status:", err)
    }

    if hasUncommitted {
        helper.SpinStopDisplay("fail")
        helper.DisplayUncommittedFiles(uncommittedFiles)
        log.Fatalln("Uncommitted files detected. Aborting workflow pause.")
    }
}
```

### Phase 4: Register Force Flag

**File**: `src/work-facilitator/cmd/pause.go`
**Location**: Lines 89-94 (init function)

**Changes**:

```go
func init() {
 rootCmd.AddCommand(pauseCmd)

 pauseCmd.Flags().StringVarP(&masterPauseArg, "master", "m", c.GOMASTER, "Go master branch rather than ref-branch")
 pauseCmd.Flags().BoolVarP(&forcePause, "force", "f", false, "Force pause workflow, skip uncommitted files check")  // NEW
}
```

### Phase 5: Update README.md

**File**: `README.md`
**Location**: After `pause` command section

**Changes**:

- Add documentation for uncommitted files detection
- Add documentation for force flag
- Include usage examples
- Note about .gitignore patterns

**Content**:

````markdown
**Uncommitted Files Detection**: The `pause` command can detect uncommitted or unstaged files before pausing a workflow. The detection respects `.gitignore` patterns (both repository-level and global). Configure behavior in `~/.workflow.yaml`:

```yaml
global:
  uncommitted_files_detection: fatal # Options: disabled, warning, fatal, interactive
```
````

- `disabled`: Skip the check
- `warning`: Show warning but continue
- `fatal`: Abort if uncommitted files found (default)
- `interactive`: Prompt for confirmation

**Force Flag**: Use `-f` or `--force` to skip the uncommitted files check:

```bash
work-facilitator pause -f
work-facilitator pause --force -m
```

**Note**: Files matching `.gitignore` patterns are automatically excluded from detection.

`````txt

## Summary of Changes

| File | Changes | Lines |
|------|---------|-------|
| `src/work-facilitator/cmd/pause.go` | Add force flag variable, PreRun check, Run re-check, flag registration | +40 approx |
| `README.md` | Add documentation for pause command detection and force flag | +15 approx |

## Testing Strategy

### Test Scenarios

1. **Normal operation (no force, fatal mode)**:

   ```bash
   # With uncommitted files
   work-facilitator pause
   # Expected: Aborts with error
    ````

2. **Force flag with uncommitted files**:

   ```bash
   # With uncommitted files
   work-facilitator pause -f
   # Expected: Shows warning, proceeds anyway
   ```

3. **With master flag**:

   ```bash
   work-facilitator pause -m -f
   # Expected: Pauses to master, skipping check
   ```

4. **Warning mode**:

   ```bash
   # Config set to warning mode
   work-facilitator pause
   # Expected: Shows warning, continues
   ```

5. **Interactive mode**:

   ```bash
   # Config set to interactive mode
   work-facilitator pause
   # Expected: Prompts user for confirmation
   ```

## Consistency with `end` Command

✅ Same force flag (`-f` / `--force`)
✅ Same configuration option (`uncommitted_files_detection`)
✅ Same detection logic (all 4 modes)
✅ Same .gitignore respect
✅ Same warning messages
✅ Same error handling patterns
✅ Same documentation style

## Potential Considerations

1. **Error Messages**: Should be tailored to "pause" context (e.g., "before pausing workflow" instead of "before ending workflow")
2. **Re-check in Run Phase**: Optional but recommended for consistency with `end` command
3. **Flag Naming**: Using `-f` for force is consistent with `commit` command's `-f` flag (though that's for different purpose)
4. **Documentation**: Should be placed in same section as `end` command for easy comparison

## Questions for User

1. Should we also add this feature to the `commit` command? (It already has a `-f` flag but for different purpose)
2. Should we add the re-check in the Run phase, or just in PreRun?
3. Any other commands that should have this feature?
`````
