# Configuration Files

This directory contains configuration files for the work-facilitator tool.

## Files

- **`.workflow.yaml`** - Main configuration file (user-specific, not committed)
- **`.workflow.dev.yaml`** - Development/example configuration
- **`.workflow.yaml.j2`** - Jinja2 template for automated configuration generation

## Synchronization Rule

**IMPORTANT**: When updating configuration structure, ALL three files must be updated to maintain consistency:

1. **`.workflow.yaml`** - Update with actual values for your environment
2. **`.workflow.dev.yaml`** - Update with example/placeholder values
3. **`.workflow.yaml.j2`** - Update with Jinja2 template variables

### Example: Adding a New Configuration Section

When adding a new configuration section (e.g., AI settings):

#### 1. Update `.workflow.yaml` (actual values)

```yaml
ai:
  enabled: True
  provider: "vertexai"
  google_project_id: "my-project"
  google_service_account_key: "/path/to/key.json"
```

#### 2. Update `.workflow.dev.yaml` (example values)

```yaml
ai:
  enabled: False
  provider: "openai"
  google_project_id: ""
  google_service_account_key: ""
```

#### 3. Update `.workflow.yaml.j2` (template variables)

```yaml
ai:
  enabled: {{ facilitators.work.ai.enabled | default(False) | quote }}
  provider: {{ facilitators.work.ai.provider | default("openai") | quote }}
  google_project_id: {{ facilitators.work.ai.google_project_id | default("") | quote }}
  google_service_account_key: {{ facilitators.work.ai.google_service_account_key | default("") | quote }}
```

## Configuration Sections

### Global Settings

- Application metadata (name, version)
- Standard enforcement
- Branch and commit patterns
- Type mappings

### Ticketing Integration

- JIRA configuration
- GitLab configuration

### AI Integration (New)

- Provider selection (openai, claude, vertexai)
- API keys and authentication
- Model configuration
- Vertex AI specific settings:
  - Google Cloud project ID
  - Region/location
  - Service account key path

## Environment Variables

Configuration values can reference environment variables using the `$VAR_NAME` syntax:

```yaml
ai:
  api_key: "$OPENAI_API_KEY"
  google_service_account_key: "$GOOGLE_SERVICE_ACCOUNT_KEY"
```

## Best Practices

1. **Never commit sensitive data** - Use environment variables for API keys and tokens
2. **Keep files in sync** - Always update all three files when changing structure
3. **Use meaningful defaults** - The `.dev.yaml` file should have safe, working defaults
4. **Document new fields** - Add comments explaining new configuration options
5. **Test templates** - Verify `.j2` template renders correctly with your variable structure

## Validation

Before committing changes to configuration files:

1. Verify all three files have the same structure
2. Check that `.dev.yaml` has safe example values
3. Ensure `.j2` template has proper Jinja2 syntax
4. Test that the application can parse all files
5. Run `make test-dev` to verify configuration loading

## Recent Changes

### 2024-12-15: Added AI Integration

- Added `ai` section to all configuration files
- Support for OpenAI, Claude, and Vertex AI providers
- Vertex AI specific settings for Google Cloud integration
