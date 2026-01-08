# Vertex AI Publisher Configuration Implementation Plan

## Objective

Enable the Vertex AI provider to parse model strings in the format `publisher/model` (e.g., `google-anthropic-claude/claude-sonnet-4.5-20250517`) and dynamically set the publisher in the API endpoint URL.

## User Requirements (Confirmed)

1. Default publisher: "google"
2. Allow any publisher string (no validation list)
3. Users specify full model string in ~/.workflow.yaml
4. Use only the first "/" as separator for multiple slashes
5. Add debug logging for parsed values
6. Export parsing function as public for other providers

## Phase 1: Code Structure Changes

### 1.1 Update VertexAIProvider Struct

**File**: `src/work-facilitator/ai/vertexai.go` (lines 32-40)

Add new field:

```go
publisher string  // stores the publisher (e.g., "google-anthropic-claude")
```

Keep model field but semantics change to store only model name.

### 1.2 Create Public Model Parsing Function

**File**: `src/work-facilitator/ai/vertexai.go` (add before NewVertexAIProvider)

Function: `ParseModelString(modelStr string) (publisher, model string)`

- **Exported** (public) for use by other providers
- Parse format: "publisher/model" or just "model"
- If no "/" found: use default publisher "google"
- If multiple "/" exist: use only first one as separator
- Add inline comments explaining the format
- Log debug info: "Parsed model string: input=%s, publisher=%s, model=%s"

Examples:

- "google-anthropic-claude/claude-sonnet-4.5-20250517" → ("google-anthropic-claude", "claude-sonnet-4.5-20250517")
- "gemini-2.5-flash" → ("google", "gemini-2.5-flash")
- "custom/model/with/slashes" → ("custom", "model/with/slashes")

### 1.3 Update NewVertexAIProvider Constructor

**File**: `src/work-facilitator/ai/vertexai.go` (lines 57-86)

Changes:

1. After model default assignment, call ParseModelString()
2. Store both publisher and parsed model in struct
3. Add debug logging: "Vertex AI provider initialized with publisher: %s model: %s"

### 1.4 Update buildEndpoint() Method

**File**: `src/work-facilitator/ai/vertexai.go` (lines 213-221)

Changes:

- Replace hardcoded "google" with p.publisher (2 occurrences)
- Update comment to reflect dynamic publisher

### 1.5 Update Validate() Method

**File**: `src/work-facilitator/ai/vertexai.go` (lines 94-122)

Changes:

- Add validation: if p.publisher == "" return error "publisher is required"

## Phase 2: Testing Strategy

### Test Cases for ParseModelString()

- With publisher: "google-anthropic-claude/claude-sonnet-4.5-20250517"
- Without publisher: "gemini-2.5-flash"
- Multiple slashes: "custom/model/with/slashes"
- Empty string: ""
- Only publisher: "publisher/"
- Custom publisher: "my-custom-publisher/my-model"

### Test Cases for NewVertexAIProvider()

- Parses publisher correctly
- Defaults publisher when not specified
- Uses default model when empty

### Test Cases for buildEndpoint()

- Custom publisher in endpoint
- Default publisher in endpoint
- Global location handling

### Test Cases for Validate()

- Publisher required validation
- Custom publisher validation

## Phase 3: Configuration & Documentation

### Configuration File Usage

Current (backward compatible):

```yaml
ai:
  provider: vertexai
  model: gemini-2.5-flash
```

New (with custom publisher):

```yaml
ai:
  provider: vertexai
  model: google-anthropic-claude/claude-sonnet-4.5-20250517
```

### Debug Logging Output

```
DEBUG: Parsed model string: input=google-anthropic-claude/claude-sonnet-4.5-20250517, publisher=google-anthropic-claude, model=claude-sonnet-4.5-20250517
DEBUG: Vertex AI provider initialized with publisher: google-anthropic-claude model: claude-sonnet-4.5-20250517
DEBUG: Vertex AI endpoint: https://us-central1-aiplatform.googleapis.com/v1/projects/my-project/locations/us-central1/publishers/google-anthropic-claude/models/claude-sonnet-4.5-20250517:generateContent
```

## Files to Modify

1. **src/work-facilitator/ai/vertexai.go**
   - Add publisher field to VertexAIProvider struct
   - Create ParseModelString() public function
   - Update NewVertexAIProvider() constructor
   - Update buildEndpoint() method
   - Update Validate() method
   - Add inline comments and debug logging

2. **src/work-facilitator/ai/vertexai_test.go**
   - Add unit tests for ParseModelString()
   - Add integration tests for NewVertexAIProvider()
   - Add endpoint generation tests
   - Add validation tests
   - Standard coverage acceptable (not 100%)

## Backward Compatibility

✅ Fully backward compatible:

- Existing model strings without "/" default to publisher "google"
- Existing ~/.workflow.yaml configurations continue to work
- No changes to public API signatures
- No changes to configuration file structure

## Implementation Order

1. Add publisher field to struct
2. Create ParseModelString() function with inline comments
3. Update NewVertexAIProvider() to use parsing
4. Update buildEndpoint() to use dynamic publisher
5. Update Validate() to check publisher
6. Add comprehensive tests
7. Verify with make build-stand-alone and make test-dev
