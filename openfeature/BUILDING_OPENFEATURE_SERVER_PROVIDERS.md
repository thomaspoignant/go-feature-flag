# Building OpenFeature Server Providers for GO Feature Flag

## Overview

This document provides language-agnostic specifications for implementing OpenFeature server-side providers for GO Feature Flag. It's based on the .NET provider implementation and serves as a comprehensive guide for developers wanting to create providers in other programming languages.

GO Feature Flag supports two evaluation modes:
- **In-Process Evaluation**: Flag configuration is cached locally and evaluation is performed using a WASM module
- **Remote Evaluation**: Each flag evaluation makes an HTTP request to the GO Feature Flag relay-proxy

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│  Application    │    │  OpenFeature     │    │  GO Feature Flag    │
│                 │────▶  Provider        │────▶  Provider           │
│                 │    │                  │    │                     │
└─────────────────┘    └──────────────────┘    └─────────────────────┘
                                                           │
                                                           ▼
                                          ┌─────────────────────────────┐
                                          │     Evaluation Strategy     │
                                          │                             │
                                          │  ┌─────────────────────────┐│
                                          │  │   In-Process            ││
                                          │  │   ┌─────────────────┐   ││
                                          │  │   │  Config Cache   │   ││
                                          │  │   └─────────────────┘   ││
                                          │  │   ┌─────────────────┐   ││
                                          │  │   │  WASM Module    │   ││
                                          │  │   └─────────────────┘   ││
                                          │  └─────────────────────────┘│
                                          │                             │
                                          │  ┌─────────────────────────┐│
                                          │  │   Remote                ││
                                          │  │   ┌─────────────────┐   ││
                                          │  │   │  HTTP Client    │   ││
                                          │  │   │  (OFREP API)    │   ││
                                          │  │   └─────────────────┘   ││
                                          │  └─────────────────────────┘│
                                          └─────────────────────────────┘
```

## Core Components

### 1. Provider Configuration Options

The provider must accept the following configuration options:

```typescript
interface ProviderOptions {
  // Required
  endpoint: string;                              // GO Feature Flag relay-proxy endpoint

  // Optional
  evaluationType?: 'InProcess' | 'Remote';       // Default: 'InProcess'
  timeout?: number;                              // HTTP timeout in milliseconds (default: 10000)
  apiKey?: string;                               // API Key for authenticated requests
  flagChangePollingIntervalMs?: number;          // Config polling interval (default: 120000)
  flushIntervalMs?: number;                      // Event collection flush interval (default: 1000)
  maxPendingEvents?: number;                     // Max events before forced flush (default: 10000)
  disableDataCollection?: boolean;               // Disable usage analytics (default: false)
  evaluationFlagList?: string[];                 // Specific flags to load (empty = all flags)
  exporterMetadata?: ExporterMetadata;           // Static metadata for analytics
  logger?: Logger;                               // Provider logger interface
  httpMessageHandler?: HttpHandler;              // Custom HTTP handler (if supported)
}

interface ExporterMetadata {
  [key: string]: string | boolean | number;      // Static metadata sent with analytics
}
```

### 2. Provider Interface Implementation

The provider must implement the OpenFeature provider interface for the target language.

## In-Process Evaluation Implementation

### Configuration Management

#### 1. Flag Configuration API

**Endpoint**: `POST {endpoint}/v1/flag/configuration`

**Headers**:
- `Content-Type: application/json`
- `X-API-Key: {apiKey}` (if configured)
- `If-None-Match: {etag}` (for caching)

**Request Body**:
```json
{
  "flags": ["flag1", "flag2"]  // Optional: specific flags to retrieve
}
```

**Response** (200 OK):
```json
{
  "flags": {
    "flag-key": {
      "variations": {
        "true": true,
        "false": false
      },
      "defaultRule": {
        "name": "legacyDefaultRule",
        "percentageRollout": {
          "true": 100,
          "false": 0
        }
      },
      "targeting": [
        {
          "name": "rule1",
          "query": "email eq \"user@example.com\"",
          "percentageRollout": {
            "true": 0,
            "false": 100
          }
        }
      ],
      "disable": false,
      "trackEvents": true,
      "version": "1.0.0"
    }
  },
  "evaluationContextEnrichment": {
    "staticKey": "staticValue"
  }
}
```
Full format details can be found in the code [here](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/flag/internal_flag.go).

**Response Headers**:
- `ETag`: Configuration version hash
- `Last-Modified`: Configuration last update time

#### 2. Configuration Polling

- Poll the configuration endpoint every `flagChangePollingIntervalMs`
- Use ETag for conditional requests to minimize bandwidth
- Compare `Last-Modified` headers to detect stale configurations
- Emit `PROVIDER_CONFIGURATION_CHANGED` events when configuration updates

### WASM Integration

#### 1. WASM Module Location

- Include the latest WASM module from GO Feature Flag releases
- Module name: `gofeatureflag-evaluation_v{version}.wasm` or `.wasi`
- Embed the module as a resource in your provider package

#### 2. WASM Runtime Setup

Required WASM functions to expose:
- `malloc(size: i32) -> i32`: Memory allocation
- `free(ptr: i32)`: Memory deallocation  
- `evaluate(inputPtr: i32, inputLen: i32) -> i64`: Flag evaluation

Required WASM memory access:
- `memory`: Linear memory for data exchange

#### 3. WASM Input Format

```json
{
  "flagKey": "my-flag",
  "flag": {
    // Complete flag configuration object, it comes from the configuration API
  },
  "evalContext": {
    "targetingKey": "user-123",
    "email": "user@example.com"
  },
  "flagContext": {
    "defaultSdkValue": false,
    "evaluationContextEnrichment": {
      "staticKey": "staticValue"
    }
  }
}
```

#### 4. WASM Output Format

```json
{
  "value": true,
  "variationType": "true",
  "reason": "TARGETING_MATCH",
  "trackEvents": true,
  "metadata": {
    "flagKey": "my-flag",
    "version": "1.0.0"
  },
  "errorCode": null,
  "errorDetails": null
}
```

#### 5. WASM Memory Management

1. **Input Processing**:
   ```
   inputJson = serialize(wasmInput)
   inputPtr = wasm.malloc(inputJson.length + 1)
   wasm.memory.writeString(inputPtr, inputJson)
   ```

2. **Evaluation**:
   ```
   result = wasm.evaluate(inputPtr, inputJson.length)
   upperBits = result >> 32      // Output pointer
   lowerBits = result & 0xFFFFFFFF  // Output length
   ```

3. **Output Processing**:
   ```
   outputJson = wasm.memory.readString(upperBits, lowerBits)
   response = deserialize(outputJson)
   ```

4. **Cleanup**:
   ```
   wasm.free(inputPtr)
   // Note: WASM module manages output memory internally
   ```

#### 6. Error Handling

Handle these WASM-specific exceptions:
- `WasmNotLoadedException`: Module failed to load
- `WasmFunctionNotFoundException`: Required function missing
- `WasmInvalidResultException`: Invalid evaluation result

### Implementation Guidelines

1. **Thread Safety**: WASM evaluation should be thread-safe or protected by synchronization
2. **Performance**: Cache WASM instances when possible; module initialization is expensive
3. **Memory**: Monitor WASM memory usage; implement cleanup strategies for long-running applications
4. **Validation**: Validate WASM output before converting to OpenFeature types

## Remote Evaluation Implementation

### OFREP Integration

Remote evaluation uses the [OpenFeature Remote Evaluation Protocol (OFREP)](https://openfeature.dev/specification/appendix-c/).

If available, you can use the OFREP implementation provided by OpenFeature. If not, you can implement the OFREP protocol directly.

#### 1. Base Configuration

**Base URL**: `{endpoint}/ofrep/v1/`

**Headers**:
- `Content-Type: application/json`
- `X-API-Key: {apiKey}` (if configured)

#### 2. Flag Evaluation Endpoint

**Endpoint**: `POST {baseUrl}/evaluate/flags/{flagKey}`

**Request Body**:
```json
{
  "context": {
    "targetingKey": "user-123",
    "email": "user@example.com"
  }
}
```

**Response**:
```json
{
  "key": "my-flag",
  "reason": "TARGETING_MATCH",
  "variant": "true",
  "value": true,
  "metadata": {
    "version": "1.0.0"
  }
}
```

### Implementation Guidelines

1. **HTTP Client**: Use language-appropriate HTTP client with timeout support
2. **Error Handling**: Map HTTP status codes to OpenFeature error types
3. **Retries**: Implement retry logic for transient failures
4. **Performance**: Consider connection pooling and keep-alive for high-throughput scenarios

## Data Collection and Analytics

### Event Publishing

#### 1. Event Types

**Feature Events**:
```json
{
  "kind": "feature",
  "creationDate": 1234567890,
  "contextKind": "user",
  "userKey": "user-123",
  "key": "my-flag",
  "variation": "true",
  "value": true,
  "default": false,
  "version": "1.0.0",
  "source": "PROVIDER"
}
```

**Tracking Events**:
```json
{
  "kind": "custom",
  "creationDate": 1234567890,
  "contextKind": "user", 
  "userKey": "user-123",
  "key": "conversion",
  "metadata": {
    "value": 29.99
  }
}
```

#### 2. Data Collector API

**Endpoint**: `POST {endpoint}/v1/data/collector`

**Request Body**:
```json
{
  "metadata": {
    "provider": "my-provider",
    "version": "1.0.0"
  },
  "events": [
    // Array of feature and tracking events
  ]
}
```

#### 3. Event Buffering

1. **Buffer Management**:
   - Maintain in-memory event buffer
   - Flush when `maxPendingEvents` reached
   - Flush periodically every `flushIntervalMs`
   - Flush on provider shutdown

2. **Error Handling**:
   - Log collection failures (don't throw)
   - Implement exponential backoff for retries
   - Drop events if buffer overflows

## Provider Hooks

### Required Hooks

#### 1. Evaluation Context Enrichment Hook

Enriches evaluation context with static metadata:

```typescript
class EnrichEvaluationContextHook implements Hook {
  constructor(private exporterMetadata: ExporterMetadata) {}

  before(hookContext: BeforeHookContext): EvaluationContext {
    const enriched = { ...hookContext.context };
    
    // Add static metadata to context
    for (const [key, value] of Object.entries(this.exporterMetadata)) {
      if (!(key in enriched)) {
        enriched[key] = value;
      }
    }
    
    return enriched;
  }
}
```

#### 2. Data Collection Hook (In-Process Only)

Collects evaluation events for analytics:

```typescript
class DataCollectorHook implements Hook {
  constructor(
    private evaluator: Evaluator,
    private eventPublisher: EventPublisher
  ) {}

  after(hookContext: AfterHookContext): void {
    if (!this.evaluator.isFlagTrackable(hookContext.flagKey)) {
      return;
    }

    const event = {
      kind: "feature",
      creationDate: Date.now() / 1000,
      contextKind: hookContext.context.anonymous ? "anonymousUser" : "user",
      userKey: hookContext.context.targetingKey || "undefined-targetingKey",
      key: hookContext.flagKey,
      variation: hookContext.evaluationDetails.variant,
      value: hookContext.evaluationDetails.value,
      default: hookContext.defaultValue,
      version: hookContext.evaluationDetails.metadata?.version,
      source: "PROVIDER"
    };

    this.eventPublisher.addEvent(event);
  }

  error(hookContext: ErrorHookContext): void {
    // Also collect error events for debugging
    this.after(hookContext);
  }
}
```

## Error Handling

### Error Type Mapping

Map GO Feature Flag errors to OpenFeature error types:

| GO Feature Flag Error | OpenFeature Error Type | Description |
|----------------------|------------------------|-------------|
| `FLAG_NOT_FOUND` | `FLAG_NOT_FOUND` | Flag doesn't exist in configuration |
| `TYPE_MISMATCH` | `TYPE_MISMATCH` | Requested type doesn't match flag type |
| `GENERAL` | `GENERAL` | Generic evaluation error |
| Network/HTTP errors | `GENERAL` | Connection or server errors |
| WASM errors | `GENERAL` | WASM runtime or evaluation errors |

### Exception Handling

Define provider-specific exceptions:

```typescript
class GoFeatureFlagException extends Error {
  constructor(message: string, cause?: Error) {
    super(message);
    this.cause = cause;
  }
}

class InvalidOptionException extends GoFeatureFlagException {}
class ImpossibleToRetrieveConfigurationException extends GoFeatureFlagException {}
class WasmNotLoadedException extends GoFeatureFlagException {}
class WasmFunctionNotFoundException extends GoFeatureFlagException {}
class WasmInvalidResultException extends GoFeatureFlagException {}
```

## Security Considerations
### Authentication

- Support API Key token authentication
- Validate API key format and presence
- Handle authentication errors gracefully

### WASM Security

- Validate WASM module signatures when possible
- Implement WASM execution timeouts
- Isolate WASM execution from host environment

### Data Privacy

- Ensure evaluation context data is handled securely
- Provide options to disable data collection
- Implement data retention policies in event buffers

## Deployment and Distribution

### Package Structure

```
provider-package/
├── src/
│   ├── provider.{ext}           # Main provider implementation
│   ├── evaluators/
│   │   ├── in-process.{ext}     # In-process evaluator
│   │   └── remote.{ext}         # Remote evaluator
│   ├── wasm/
│   │   ├── evaluator.{ext}      # WASM integration
│   │   └── gofeatureflag.wasm   # Embedded WASM module
│   ├── api/
│   │   └── client.{ext}         # HTTP API client
│   └── models/
│       └── *.{ext}              # Data models
├── tests/
├── docs/
└── README.md
```

### Dependencies

Minimize external dependencies:
- HTTP client library
- JSON serialization library
- WASM runtime (for in-process evaluation)
- OpenFeature SDK for target language

### Versioning

- Follow semantic versioning
- Maintain compatibility matrix with GO Feature Flag versions
- Document breaking changes clearly

## Conclusion

This specification provides the foundation for implementing robust OpenFeature providers for GO Feature Flag across different programming languages. Key implementation priorities:

1. **Start Simple**: Begin with remote evaluation for faster implementation
2. **Add WASM Support**: Implement in-process evaluation for better performance
3. **Focus on Reliability**: Implement proper error handling and recovery
4. **Optimize Performance**: Profile and optimize hot paths
5. **Test Thoroughly**: Ensure compatibility with GO Feature Flag versions

For questions or clarifications, refer to the GO Feature Flag documentation or examine the reference .NET implementation in this repository.
