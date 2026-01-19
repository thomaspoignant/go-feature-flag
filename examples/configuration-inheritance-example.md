# Configuration Inheritance Example

This example demonstrates how flagSets can inherit configuration from the top-level configuration while still being able to override specific values.

## Example Configuration

```yaml
# Top-level configuration - these values are inherited by all flagsets
retriever:
  kind: http
  url: https://example.com/common-flags.yaml
exporter:
  kind: log
fileFormat: yaml
pollingInterval: 120
environment: production
enablePollingJitter: true
evaluationContextEnrichment:
  serverVersion: "1.0.0"
  region: "us-west-2"

# FlagSets configuration
flagsets:
  # This flagset inherits all top-level settings but overrides specific values
  - name: "frontend-flagset"
    apiKeys:
      - "frontend-api-key-1"
      - "frontend-api-key-2"
    # Override only the retriever, inherit everything else
    retrievers:
      - kind: s3
        bucket: "frontend-flags"
        item: "feature-flags.yaml"
    # Override polling interval to be more frequent for frontend
    pollingInterval: 30
    # Inherits: exporter (log), fileFormat (yaml), environment (production), etc.

  # This flagset inherits all top-level settings without overrides
  - name: "backend-flagset"
    apiKeys:
      - "backend-api-key-1"
    # Only provide API keys, inherit everything else from top-level
    # Inherits: retriever (http), exporter (log), fileFormat (yaml), 
    #          pollingInterval (120), environment (production), etc.

  # This flagset overrides multiple values
  - name: "admin-flagset"
    apiKeys:
      - "admin-api-key"
    # Override retriever
    retrievers:
      - kind: github
        repository: "company/feature-flags"
        path: "admin-flags.yaml"
    # Override exporter
    exporters:
      - kind: webhook
        endpointUrl: "https://hooks.slack.com/admin-webhook"
    # Override environment
    environment: staging
    # Inherits: fileFormat (yaml), pollingInterval (120), etc.
```

## What Gets Inherited

The following fields from `CommonFlagSet` are inherited from top-level configuration:

- `Retriever` / `Retrievers`
- `Notifiers`
- `Exporter` / `Exporters`
- `FileFormat`
- `PollingInterval`
- `StartWithRetrieverError`
- `EnablePollingJitter`
- `DisableNotifierOnInit`
- `EvaluationContextEnrichment`
- `PersistentFlagConfigurationFile`
- `Environment`

## Inheritance Rules

1. **Flagset values take precedence**: If a field is defined in a flagset, it uses that value
2. **Top-level values are inherited**: If a field is not defined in a flagset, it uses the top-level value
3. **No inheritance occurs**: If neither flagset nor top-level defines a field, it uses the default value

## Benefits

- **DRY principle**: Common settings are defined once at the top level
- **Flexibility**: Flagsets can override specific settings when needed
- **Maintainability**: Changes to common settings affect all flagsets unless explicitly overridden
- **Backward compatibility**: Existing configurations continue to work unchanged

## Use Cases

1. **Common exporters**: All flagsets use the same exporter (e.g., log) but different retrievers
2. **Common polling intervals**: Most flagsets use the same polling interval, with exceptions
3. **Common environment**: All flagsets run in the same environment unless specified otherwise
4. **Common evaluation context**: Add common attributes to all evaluations (server version, region, etc.)