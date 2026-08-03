---
title: GO Feature Flag Provider Specification
---

# Specification Document for GO Feature Flag OpenFeature Providers

|                             |                                          |
| --------------------------- | ---------------------------------------- |
| **Specification Version**   | 1.0                                      |
| **Creation Date**           | 03/08/2026                               |
| **Last Update Date**        | 03/08/2026                               |
| **Authors**                 | Thomas Poignant                          |
| **Minimum relay proxy**     | `v1.55.0`                                |
| **Evaluation engine**       | `modules/core v0.7.2` (WASM `0.2.3`)     |

## Overview

GO Feature Flag does not ship its own SDKs. Applications talk to it through
[OpenFeature](https://openfeature.dev) SDKs plus a GO Feature Flag **provider**. Those
providers exist in eight languages, written by different people at different times, and they
have drifted apart — not only in naming and defaults, but in wire format, error semantics and
evaluation results.

This document is the contract every server-side GO Feature Flag provider must satisfy so that
**the same flag, evaluated with the same context, returns the same result in every language.**

It is written to be checked mechanically. Every normative statement carries a stable
identifier, an RFC 2119 keyword and a severity, so an automated audit can walk a provider's
source and emit a per-requirement verdict.

:::info
This specification covers **server-side** providers only. Client-side providers (Swift,
Android, JavaScript Web) follow a different paradigm and are out of scope for version 1.0.
:::

### Relationship to other documents

- It **supersedes** `BUILDING_OPENFEATURE_SERVER_PROVIDERS.md`, which was written from the
  .NET provider and is inaccurate in several load-bearing places.
- It **absorbs** the former Provider Cache specification, now [§17](#17-remote-cache-optional).

---

## 1. Scope and conventions

### 1.1 Requirement keywords

The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHOULD**, **SHOULD NOT** and **MAY** are
to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

### 1.2 Requirement identifiers

Every requirement has a stable identifier of the form `GOFF-<AREA>-<NNN>`. Identifiers are
never reused or renumbered; a withdrawn requirement is marked as such and its number retired.

### 1.3 Conformance tiers

Not every requirement applies to every provider. Each section declares a tier:

| Tier            | Applies to                                                          |
| --------------- | ------------------------------------------------------------------- |
| **Core**        | Every server-side provider                                          |
| **Remote**      | Providers offering remote (OFREP) evaluation                        |
| **In-process**  | Providers offering local evaluation                                 |
| **WASM**        | In-process providers that embed the engine through WebAssembly      |
| **Optional**    | MAY-level capabilities                                              |

A requirement outside a provider's tiers is reported **N/A**, never **FAIL**. In particular, a
provider written in Go links the evaluation engine directly and is **N/A** for the whole of
[§10](#10-webassembly-abi--goff-wasm) — it does not fail for declining to use WebAssembly.

Where the target language's OpenFeature SDK offers no equivalent capability at all — for
example, an SDK with no Tracking API — the corresponding requirements are also **N/A**.

### 1.4 Severity

| Severity     | Meaning                                                                 |
| ------------ | ----------------------------------------------------------------------- |
| **Critical** | Wrong flag values, silent data loss, or unrecoverable provider state    |
| **Major**    | Observable misbehaviour, or a contracted capability that is missing     |
| **Minor**    | Telemetry fidelity, naming, cosmetics                                   |

### 1.5 Deferring to the language SDK

Where the target language's OpenFeature SDK defines a behaviour — the canonical object type,
the `ErrorCode` enumeration, provider status transitions, event types, hook stage ordering —
the provider **MUST** follow its SDK. Those SDKs are implementations of one shared OpenFeature
specification, so deferring to them converges rather than diverges.

Where a divergence is a **language accident** rather than an SDK decision, this specification
overrides it. Known accidents, called out so implementers know the override is deliberate:

- **Python**: `isinstance(True, int)` is `True`, so a boolean silently satisfies an integer
  resolver. See `GOFF-EVAL-004`.
- **Java**: a JSON number above `Integer.MAX_VALUE` decodes to `Long` and currently fails both
  the integer and the double resolver.
- **.NET**: `GetInt32()` throws `FormatException` on a non-integral number rather than
  producing a type mismatch.

### 1.6 Evaluation engine version

All providers **MUST** evaluate using the same engine build. Two providers on different engine
versions can return different values for the same flag and context — a divergence no other
requirement in this document would detect, because all of them would still pass.

| ID              | Sev      | Requirement                                                                                                                                                              |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-ENG-001`  | Critical | The provider **MUST** evaluate using engine `modules/core v0.7.2`. A WASM-based provider satisfies this by pinning WASM module `0.2.3`, which embeds that core version.  |
| `GOFF-ENG-002`  | Major    | The pinned engine version **MUST** be recorded in a single machine-readable location in the provider repository (a version file, build property or dependency manifest). |
| `GOFF-ENG-003`  | Minor    | The provider **SHOULD** document which specification version it targets.                                                                                                 |

---

## 2. Provider identity — `GOFF-META`

**Tier: Core**

| ID              | Sev   | Requirement                                                                                                                                            |
| --------------- | ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-META-001` | Minor | The provider metadata name **MUST** be exactly `GO Feature Flag Provider`.                                                                              |
| `GOFF-META-002` | Minor | The metadata name **MUST** be a literal constant. It **MUST NOT** be derived by runtime reflection on a class or type name, which is unstable under minification and obfuscation. |
| `GOFF-META-003` | Minor | Any provider name carried on emitted provider events **MUST** equal the metadata name.                                                                  |

---

## 3. Configuration — `GOFF-CFG`

**Tier: Core**

Option **names** are RECOMMENDED, not normative: a provider **SHOULD** use the canonical name
adapted to its language's casing convention, and MAY retain an existing name as an alias.
Option **semantics and default values are normative** — they have real operational
consequences, and divergent defaults are how a fleet ends up behaving inconsistently.

### 3.1 Canonical options

| Canonical name                | Type        | Default          | Applies to  |
| ----------------------------- | ----------- | ---------------- | ----------- |
| `endpoint`                    | URL string  | *(required)*     | both        |
| `evaluationType`              | enum        | in-process       | both        |
| `apiKey`                      | string      | none             | both        |
| `timeout`                     | duration    | `10000 ms`       | both        |
| `flagChangePollingInterval`   | duration    | `120000 ms`      | in-process  |
| `evaluationFlagList`          | string list | empty (all)      | in-process  |
| `exporterMetadata`            | map         | empty            | both        |
| `dataFlushInterval`           | duration    | `60000 ms`       | both        |
| `maxPendingEvents`            | integer     | `10000`          | both        |
| `disableDataCollection`       | boolean     | `false`          | both        |
| `dataCollectorBaseURL`        | URL string  | `endpoint`       | both        |
| `wasmEvaluatorPoolSize`       | integer     | CPU core count   | WASM        |
| `logger`                      | SDK logger  | language default | both        |

### 3.2 Requirements

| ID             | Sev      | Requirement                                                                                                                                                                                        |
| -------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-CFG-001` | Major    | `endpoint` **MUST** be required and validated at construction time. An absent or malformed value **MUST** raise a configuration error before any network activity.                                  |
| `GOFF-CFG-002` | Major    | The default evaluation mode **MUST** be in-process.                                                                                                                                                |
| `GOFF-CFG-003` | Major    | Every option listed in §3.1 that the provider supports **MUST** use the default value given there.                                                                                                 |
| `GOFF-CFG-004` | Major    | The provider **MUST NOT** mutate the caller's options object or any collection it contains. Normalisation **MUST** operate on a copy.                                                               |
| `GOFF-CFG-005` | Critical | The provider **MUST NOT** read environment variables to determine the endpoint or credentials. A feature-flag provider silently retargeting itself based on ambient environment is a security surprise. |
| `GOFF-CFG-006` | Major    | `dataCollectorBaseURL` **MUST** override the base URL for the data-collector endpoint **only**. Flag-configuration and evaluation requests **MUST** continue to use `endpoint`. When unset it **MUST** fall back to `endpoint`. |
| `GOFF-CFG-007` | Major    | `dataCollectorBaseURL` **MUST** replace the whole base, including scheme, host, port and path prefix, and authentication, custom headers and timeout **MUST** apply to it identically.               |
| `GOFF-CFG-008` | Minor    | `evaluationFlagList` **SHOULD** be supported, and when non-empty **MUST** be transmitted as the `flags` array of the flag-configuration request.                                                    |
| `GOFF-CFG-009` | Major    | An option that the provider declares and documents **MUST** be honoured. A declared option that is never read is a defect regardless of its default.                                                |
| `GOFF-CFG-010` | Minor    | The provider **MUST NOT** expose options for capabilities it does not implement. Vestigial options from removed features **MUST** be deleted rather than left inert.                                |

---

## 4. Lifecycle — `GOFF-LIFE`

**Tier: Core**

| ID              | Sev      | Requirement                                                                                                                                                                    |
| --------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-LIFE-001` | Major    | Initialization **MUST** block until the provider can serve evaluations, or fail.                                                                                                |
| `GOFF-LIFE-002` | Critical | Initialization **MUST** be safe to call more than once. A second call **MUST** cancel and join any existing polling task, **MUST NOT** start a duplicate one, and **MUST NOT** leak or double-instantiate the evaluation engine. |
| `GOFF-LIFE-003` | Critical | Any one-shot shutdown guard **MUST** be reset by re-initialization, so that a subsequent shutdown cannot hang or panic.                                                          |
| `GOFF-LIFE-004` | Major    | Shutdown **MUST** stop the polling task, **MUST** flush all buffered events, and **MUST** stop the event publisher — in **both** evaluation modes and regardless of whether data collection is enabled. |
| `GOFF-LIFE-005` | Major    | Shutdown **MUST** bound how long it waits for background work.                                                                                                                 |
| `GOFF-LIFE-006` | Critical | Until a flag configuration has been successfully loaded at least once, evaluations **MUST** report `PROVIDER_NOT_READY`. They **MUST NOT** report `FLAG_NOT_FOUND`, which misattributes an infrastructure failure to the caller's flag key. |
| `GOFF-LIFE-007` | Major    | Concurrent evaluation **MUST** be safe. Shared configuration state **MUST** be guarded, and the guard **MUST NOT** be held across network or evaluation calls.                  |

---

## 5. Provider events and status — `GOFF-EVT`

**Tier: Core**

| ID             | Sev      | Requirement                                                                                                                                                                              |
| -------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-EVT-001` | Major    | Provider events **MUST** be emitted identically in both evaluation modes. A capability present in one mode and silently absent in the other is a defect.                                  |
| `GOFF-EVT-002` | Major    | `PROVIDER_CONFIGURATION_CHANGED` **MUST** be emitted when a poll yields a configuration different from the one in use.                                                                    |
| `GOFF-EVT-003` | Major    | `PROVIDER_CONFIGURATION_CHANGED` **MUST NOT** be emitted for the initial load during initialization. Consumers **MUST NOT** observe a configuration-changed event before the provider is ready. |
| `GOFF-EVT-004` | Major    | `PROVIDER_CONFIGURATION_CHANGED` **MUST NOT** be emitted when the configuration is unchanged. A provider that cannot distinguish "changed" from "fetched" — for instance because the server sends no `ETag` — **MUST** compare content rather than emit unconditionally. |
| `GOFF-EVT-005` | Major    | The provider **SHOULD** emit `PROVIDER_STALE` after **3** consecutive failed configuration refreshes, and **MUST** continue serving the last known-good configuration.                    |
| `GOFF-EVT-006` | Major    | On recovery from stale, the provider **MUST** emit an event returning it to ready.                                                                                                       |
| `GOFF-EVT-007` | Critical | Authentication failure (`401`, `403`) during initialization **MUST** put the provider in `PROVIDER_FATAL`. Credentials cannot be repaired by retrying.                                    |
| `GOFF-EVT-008` | Major    | Every other initialization failure **MUST** put the provider in `ERROR`, not `PROVIDER_FATAL`, so it recovers unattended once the relay proxy is reachable.                               |

---

## 6. Evaluation API — `GOFF-EVAL`

**Tier: Core**

| ID              | Sev      | Requirement                                                                                                                                                                        |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-EVAL-001` | Major    | The provider **MUST** implement the boolean, string, integer, float and object resolvers defined by its SDK, and **SHOULD** implement asynchronous variants where the SDK defines them. |
| `GOFF-EVAL-002` | Critical | The object resolver **MUST** accept exactly what its SDK's canonical structure type can represent, and **MUST** report `TYPE_MISMATCH` for anything it cannot — including scalars.  |
| `GOFF-EVAL-003` | Critical | The float resolver **MUST** accept an integral JSON number. JSON does not distinguish `100` from `100.0`.                                                                            |
| `GOFF-EVAL-004` | Critical | A boolean value **MUST NOT** satisfy the integer or float resolver. It **MUST** report `TYPE_MISMATCH`.                                                                              |
| `GOFF-EVAL-005` | Major    | The integer resolver **MUST** report `TYPE_MISMATCH` for a non-integral number. It **MUST NOT** truncate, round, or raise a raw numeric-conversion error.                            |
| `GOFF-EVAL-006` | Critical | A `null` evaluation result **MUST** return the caller's default value, preserving the engine's reason, variant and metadata. It **MUST NOT** return the language's zero value.       |
| `GOFF-EVAL-007` | Major    | The `reason` **MUST** be passed through as an opaque string. The provider **MUST NOT** parse it into a closed enumeration. The engine emits `TARGETING_MATCH_SPLIT`, `SPLIT`, `OFFLINE` and others that a naive enum lookup will reject. |
| `GOFF-EVAL-008` | Major    | A disabled flag **MUST** return the caller's default value with reason `DISABLED` and variant `SdkDefault`.                                                                          |
| `GOFF-EVAL-009` | Major    | Flag metadata **MUST** be passed through with its structure intact. Values **MUST NOT** be coerced to strings.                                                                       |
| `GOFF-EVAL-010` | Major    | Metadata keys added by the relay proxy, such as `gofeatureflag_cacheable`, **MUST** be passed through verbatim. The provider **MUST NOT** strip them, and **MUST NOT** require them. |
| `GOFF-EVAL-011` | Major    | On any evaluation error the provider **MUST** return the caller's default value together with the error code. It **MUST NOT** propagate an exception to the application.             |

---

## 7. Evaluation context — `GOFF-CTX`

**Tier: Core**

| ID             | Sev      | Requirement                                                                                                                                                                        |
| -------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-CTX-001` | Major    | The targeting key **MUST** be transmitted under the key `targetingKey`.                                                                                                             |
| `GOFF-CTX-002` | Major    | Context attributes **MUST** be flattened alongside `targetingKey`, not nested under a wrapper.                                                                                       |
| `GOFF-CTX-003` | Critical | A missing or empty targeting key **MUST** be passed through to the evaluation engine. The provider **MUST NOT** reject it. The engine returns `TARGETING_KEY_MISSING` only for flags that actually require bucketing; rejecting client-side breaks flags that do not. |
| `GOFF-CTX-004` | Critical | In-process evaluation **MUST** normalise attributes exactly as the engine's own entry point does, including narrowing integral floating-point values to integers. Skipping this makes targeting rules match differently between languages for identical input. |
| `GOFF-CTX-005` | Major    | `evaluationContextEnrichment` from the flag-configuration response **MUST** be merged into the evaluation context. On key collision, **enrichment wins**.                            |

### 7.1 The `gofeatureflag` reserved namespace

The `gofeatureflag` context key is a **shared namespace** with three fields:

| Field              | Written by | Purpose                                                     |
| ------------------ | ---------- | ----------------------------------------------------------- |
| `exporterMetadata` | provider   | Static metadata attached to exported evaluation events      |
| `flagList`         | caller     | Restricts which flags a bulk evaluation returns             |
| `currentDateTime`  | caller     | Overrides evaluation time, for testing scheduled rollouts   |

| ID             | Sev      | Requirement                                                                                                                                                                       |
| -------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-CTX-006` | Critical | The provider **MUST** merge into `gofeatureflag`, setting or replacing only `exporterMetadata` and preserving every sibling key. Replacing the whole object silently destroys caller-supplied `flagList` and `currentDateTime`. |
| `GOFF-CTX-007` | Major    | `exporterMetadata` **MUST** be nested under `gofeatureflag.exporterMetadata`. Writing the metadata flat under `gofeatureflag` means the server never reads it.                     |
| `GOFF-CTX-008` | Major    | If `gofeatureflag` is present but not a map, the provider **MUST** replace it rather than fail.                                                                                    |
| `GOFF-CTX-009` | Major    | The provider **MUST NOT** write `flagList` or `currentDateTime`. They are caller inputs.                                                                                           |

---

## 8. Remote evaluation — `GOFF-REM`

**Tier: Remote**

Remote evaluation uses the
[OpenFeature Remote Evaluation Protocol](https://openfeature.dev/specification/appendix-c/).

| ID             | Sev      | Requirement                                                                                                                                                     |
| -------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-REM-001` | Major    | Single-flag evaluation **MUST** be `POST {endpoint}/ofrep/v1/evaluate/flags/{flagKey}`.                                                                          |
| `GOFF-REM-002` | Major    | The request body **MUST** be `{"context": { ... }}` containing `targetingKey` and the flattened attributes.                                                      |
| `GOFF-REM-003` | Major    | The response fields `value`, `reason`, `variant` and `metadata` **MUST** be mapped to the SDK's resolution details. `variant` on the wire corresponds to `variationType` in-process — they are the same concept. |
| `GOFF-REM-004` | Major    | The provider **MUST** honour `429` responses by respecting `Retry-After` before issuing further requests.                                                        |
| `GOFF-REM-005` | Major    | The provider **MUST NOT** implement any other retry behaviour. Polling and the caller's own retry policy are the recovery mechanisms.                            |
| `GOFF-REM-006` | Major    | The configured `timeout` **MUST** apply to every remote request. The provider **MUST NOT** delegate to a transitive library's default.                           |

---

## 9. In-process evaluation — `GOFF-IP`

**Tier: In-process**

In-process evaluation is defined by **behaviour**, not by transport. A provider may embed the
engine through WebAssembly or, where the language permits, link it directly. Both **MUST**
produce identical results.

### 9.1 Flag configuration retrieval

| ID            | Sev      | Requirement                                                                                                                                          |
| ------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-IP-001` | Major    | The provider **MUST** fetch configuration with `POST {endpoint}/v1/flag/configuration`.                                                               |
| `GOFF-IP-002` | Major    | The request body **MUST** be `{"flags": [...]}`, an empty array meaning "all flags".                                                                  |
| `GOFF-IP-003` | Critical | Any path prefix on `endpoint` **MUST** be preserved. Building the URL from an absolute path discards it and silently retargets the request.           |
| `GOFF-IP-004` | Major    | The response `flags` and `evaluationContextEnrichment` **MUST** both be stored.                                                                       |
| `GOFF-IP-005` | Major    | The `ETag` response header **MUST** be stored verbatim, including surrounding quotes, and echoed verbatim as `If-None-Match`. The relay proxy issues strong validators; stripping quotes breaks the comparison. |

### 9.2 Polling and refresh

| ID            | Sev      | Requirement                                                                                                                                                                                     |
| ------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-IP-006` | Critical | The provider **MUST** poll for configuration changes on the configured interval. Polling **MUST** be active by default; it **MUST NOT** require explicit opt-in.                                 |
| `GOFF-IP-007` | Critical | A `304 Not Modified` response **MUST** be treated as "no change". It **MUST NOT** write flags, enrichment, timestamps or the stored `ETag` — **regardless of whether the response echoed an `ETag` header**. |
| `GOFF-IP-008` | Critical | A `200` response whose body cannot be parsed **MUST** be treated as a failed refresh: the previous configuration **MUST** be preserved and the stored `ETag` **MUST NOT** advance.               |
| `GOFF-IP-009` | Critical | A `200` response whose decoded flag map is null or absent **MUST** likewise be treated as a failed refresh. Accepting it wipes every flag, and advancing the `ETag` makes the empty state permanent. |
| `GOFF-IP-010` | Major    | A failed refresh **MUST NOT** terminate polling. Polling **MUST** survive any error and continue on schedule.                                                                                    |
| `GOFF-IP-011` | Minor    | The provider **SHOULD** apply jitter to the polling interval so that a restarted fleet does not poll in lockstep.                                                                                |

:::tip Recommended implementation
Make `GOFF-IP-007` correct by construction rather than by null-checking. Have the HTTP layer
return a dedicated "not modified" sentinel *instead of* a response object, so the 304 branch is
structurally incapable of carrying a parseable body, and return from the refresh routine before
acquiring any lock on the configuration state.
:::

### 9.3 Evaluation

| ID            | Sev      | Requirement                                                                                                                                                     |
| ------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-IP-012` | Major    | A flag absent from the local configuration **MUST** yield `FLAG_NOT_FOUND` without invoking the engine.                                                          |
| `GOFF-IP-013` | Critical | Evaluation **MUST** be panic- and exception-safe. An engine fault **MUST** degrade to a `GENERAL` error and the caller's default value; it **MUST NOT** propagate into the application. |
| `GOFF-IP-014` | Major    | The engine input **MUST** carry `flagKey`, `flag`, `evalContext`, and `flagContext` containing `defaultSdkValue` and `evaluationContextEnrichment`.              |

---

## 10. WebAssembly ABI — `GOFF-WASM`

**Tier: WASM.** Providers that link the engine directly are **N/A** for this section.

The module exposes one entry point. The host writes a JSON request into linear memory, calls
`evaluate`, and reads a JSON response back out.

```text
1. Serialize the input to UTF-8 bytes
2. ptr = malloc(byteLength + 1)
3. write bytes at ptr, followed by a NUL terminator
4. packed = evaluate(ptr, byteLength)      // note: byteLength, NOT including the terminator
5. free(ptr)
6. outPtr = (packed >> 32) & 0xFFFFFFFF
   outLen =  packed        & 0xFFFFFFFF
7. read outLen bytes at outPtr and parse as JSON
```

| ID              | Sev      | Requirement                                                                                                                                                                       |
| --------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-WASM-001` | Major    | The host **MUST** resolve the exports `memory`, `malloc`, `free` and `evaluate`, and **MUST** fail initialization if any is absent.                                                |
| `GOFF-WASM-002` | Major    | If `_start` is exported it **MUST** be invoked once per instance, and an exit code of `0` **MUST** be tolerated rather than treated as failure.                                    |
| `GOFF-WASM-003` | Critical | The length passed to `evaluate` **MUST** be the **UTF-8 byte length** of the serialized input. Passing a string length measured in UTF-16 code units truncates the payload for any non-ASCII input. |
| `GOFF-WASM-004` | Major    | The packed `i64` result **MUST** be unpacked as pointer in the high 32 bits and length in the low 32 bits, using arithmetic wide enough not to overflow.                           |
| `GOFF-WASM-005` | Major    | The host **MUST** free the input pointer, and **MUST NOT** free the output pointer — the guest owns the output buffer.                                                             |
| `GOFF-WASM-006` | Major    | A zero pointer or zero length **MUST** be treated as an invalid result.                                                                                                           |
| `GOFF-WASM-007` | Critical | Concurrent evaluation **MUST** be safe. A single instance **MUST NOT** be shared across threads without serialisation; a pool of instances is the RECOMMENDED approach.            |
| `GOFF-WASM-008` | Critical | If evaluation traps, the instance **MUST** be discarded and rebuilt. A trapped instance has undefined linear memory and **MUST NOT** be returned to a pool or reused.              |
| `GOFF-WASM-009` | Major    | The instance pool **SHOULD** default to the host's CPU core count and **SHOULD** be configurable.                                                                                 |
| `GOFF-WASM-010` | Major    | The engine binary version **MUST** be pinned to a single, machine-readable value.                                                                                                 |
| `GOFF-WASM-011` | Minor    | The provider **SHOULD** allow the binary path to be overridden, so that bundlers and non-standard packaging layouts remain usable.                                                 |

Input and output shapes are given in [Appendix B.3](#b3-engine-abi-vectors).

---

## 11. Error model — `GOFF-ERR`

**Tier: Core**

### 11.1 Engine error codes

`PROVIDER_NOT_READY`, `FLAG_NOT_FOUND`, `PARSE_ERROR`, `TYPE_MISMATCH`, `GENERAL`,
`INVALID_CONTEXT`, `TARGETING_KEY_MISSING`, and the GO Feature Flag-specific `FLAG_CONFIG`.

### 11.2 Requirements

| ID             | Sev      | Requirement                                                                                                                                            |
| -------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-ERR-001` | Major    | Engine error codes with an SDK equivalent **MUST** be mapped to it.                                                                                    |
| `GOFF-ERR-002` | Major    | Any unrecognised error code **MUST** map to `GENERAL`. It **MUST NOT** map to null, and **MUST NOT** raise an unmapped language-level exception.        |
| `GOFF-ERR-003` | Major    | `errorDetails` **MUST** be carried through as the error message.                                                                                       |
| `GOFF-ERR-004` | Major    | HTTP failures **MUST** be distinguishable by status. `401` and `403` **MUST** be reported distinctly from `404`, `400`, `429` and `5xx`.                |
| `GOFF-ERR-005` | Major    | Errors raised during background refresh **MUST** be logged and **MUST NOT** propagate to the application.                                              |
| `GOFF-ERR-006` | Minor    | Data-collector failures **MUST** be logged. Silently discarding them makes a permanently failing exporter undetectable.                                 |

---

## 12. Hooks — `GOFF-HOOK`

**Tier: Core**

| ID              | Sev      | Requirement                                                                                                                                                     |
| --------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-HOOK-001` | Major    | The provider's hooks **MUST** be observable by the time initialization completes, in the order **`[EnrichEvaluationContext, DataCollector]`**.                   |
| `GOFF-HOOK-002` | Major    | Repeated initialization **MUST NOT** duplicate hooks.                                                                                                           |
| `GOFF-HOOK-003` | Major    | The enrichment hook **MUST** be registered unconditionally. Because `exporterMetadata` always contains the reserved keys of `GOFF-COLL-010`, it always has something to contribute. |
| `GOFF-HOOK-004` | Major    | The enrichment hook **MUST** implement only the `before` stage, and **MUST** return a new context rather than mutating the caller's.                             |
| `GOFF-HOOK-005` | Major    | The data-collector hook **MUST** implement the `after` and `error` stages.                                                                                       |
| `GOFF-HOOK-006` | Major    | Both stages **MUST** honour `disableDataCollection` and the flag's trackability. Gating only one stage produces partial telemetry that looks like data loss.     |

---

## 13. Data collection — `GOFF-COLL`

**Tier: Core**

Evaluation and tracking events are batched and posted to the relay proxy.

### 13.1 Envelope

```json
{
  "meta": { "provider": "python", "openfeature": true },
  "events": [ ... ]
}
```

| ID              | Sev      | Requirement                                                                                                                                                          |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-COLL-001` | Major    | Events **MUST** be posted to `POST {dataCollectorBaseURL}/v1/data/collector`.                                                                                          |
| `GOFF-COLL-002` | Critical | The metadata key **MUST** be `meta`. The events key **MUST** be `events`.                                                                                              |

### 13.2 Feature event

| ID              | Sev      | Requirement                                                                                                                                                          |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-COLL-003` | Major    | `kind` **MUST** be `feature`.                                                                                                                                         |
| `GOFF-COLL-004` | Critical | The "evaluation failed" boolean **MUST** be serialised as `default`. Any other name is silently discarded by the relay proxy, recording every failed evaluation as a success. |
| `GOFF-COLL-005` | Major    | `contextKind` **MUST** be `user`, unless the context carries `anonymous` explicitly set to `true`, in which case it **MUST** be `anonymousUser`. An absent context **MUST** yield `user`. |
| `GOFF-COLL-006` | Major    | `userKey` **MUST** be the targeting key, or the sentinel `undefined-targetingKey` when absent.                                                                         |
| `GOFF-COLL-007` | Major    | `creationDate` **MUST** be Unix epoch **seconds**.                                                                                                                    |
| `GOFF-COLL-008` | Major    | `variation` **MUST** be the resolved variant, or `SdkDefault` when none is available.                                                                                 |
| `GOFF-COLL-009` | Minor    | `version` **MUST** be populated from flag metadata when present.                                                                                                      |
| `GOFF-COLL-010` | Minor    | `source` **MUST** be `INPROCESS` for a locally evaluated flag, or `PROVIDER_CACHE` for a value served from a remote-mode cache. `SERVER` is reserved for the relay proxy. |

### 13.3 Exporter metadata

| ID              | Sev      | Requirement                                                                                                                                                          |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-COLL-011` | Major    | The `meta` envelope **MUST** always contain `provider` and `openfeature: true`, whether or not the user configured any metadata. Without them events cannot be attributed to an SDK. |
| `GOFF-COLL-012` | Minor    | `provider` **MUST** be the lowercase language identifier: `python`, `java`, `dotnet`, `go`, `javascript`, `kotlin`, `php`, `ruby`, `rust`.                             |
| `GOFF-COLL-013` | Major    | `exporterMetadata` values **MUST** be restricted to string, boolean, integer or floating-point, and an invalid value **MUST** be rejected at construction time.        |

### 13.4 Buffering and flushing

| ID              | Sev      | Requirement                                                                                                                                                          |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-COLL-014` | Major    | Events **MUST** be flushed on the configured interval, when the buffer reaches `maxPendingEvents`, and on shutdown.                                                    |
| `GOFF-COLL-015` | Minor    | The publisher **MUST NOT** flush immediately on start. There is nothing to send.                                                                                       |
| `GOFF-COLL-016` | Major    | Flushing **MUST** be single-flight — concurrent publishes **MUST NOT** overlap.                                                                                        |
| `GOFF-COLL-017` | Critical | Single-flight **MUST NOT** be achieved by holding a lock across the HTTP call. Enqueuing an event runs inside an evaluation hook; blocking it couples flag-evaluation latency to data-collector availability. Swap the buffer under the lock, release, then post. |
| `GOFF-COLL-018` | Major    | A failed batch **MUST** be re-queued preserving chronological order.                                                                                                   |
| `GOFF-COLL-019` | Critical | The buffer **MUST** be capped at twice `maxPendingEvents`, discarding **oldest** events on overflow. An uncapped buffer is an unbounded memory leak during a collector outage. |

### 13.5 What is collected

| ID              | Sev      | Requirement                                                                                                                                                          |
| --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `GOFF-COLL-020` | Major    | A flag whose configuration omits `trackEvents` **MUST** be treated as trackable. The engine's own default is `true`.                                                   |
| `GOFF-COLL-021` | Major    | A flag absent from the configuration **MUST** be treated as trackable, so that a flag added between polls still produces data.                                         |
| `GOFF-COLL-022` | Major    | A flag whose configuration sets `trackEvents: false` **MUST NOT** produce feature events.                                                                              |
| `GOFF-COLL-023` | Major    | In remote mode, feature events **MUST** be emitted only for evaluations served from a provider cache. Uncached remote evaluations are already recorded by the relay proxy and **MUST NOT** be counted twice. |

---

## 14. Tracking — `GOFF-TRACK`

**Tier: Core.** N/A where the language's OpenFeature SDK has no Tracking API.

| ID               | Sev   | Requirement                                                                                                                       |
| ---------------- | ----- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-TRACK-001` | Major | The provider **MUST** implement the SDK's Tracking API where one exists.                                                          |
| `GOFF-TRACK-002` | Major | Tracking events **MUST** be sent in **both** evaluation modes. The relay proxy does not synthesise custom events.                 |
| `GOFF-TRACK-003` | Major | Tracking events **MUST** honour `disableDataCollection`.                                                                          |
| `GOFF-TRACK-004` | Major | `kind` **MUST** be `tracking`, and the details field **MUST** be `trackingEventDetails`.                                          |
| `GOFF-TRACK-005` | Major | The event **MUST** carry `evaluationContext`, and **MUST** use the same `contextKind`, `userKey` and `creationDate` rules as §13. |

---

## 15. Authentication — `GOFF-AUTH`

**Tier: Core**

| ID              | Sev      | Requirement                                                                                                                                  |
| --------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-AUTH-001` | Major    | When `apiKey` is set, the provider **MUST** send `Authorization: Bearer {apiKey}`.                                                            |
| `GOFF-AUTH-002` | Major    | The header **MUST** be applied to every authenticated endpoint: flag configuration, evaluation and data collection.                           |
| `GOFF-AUTH-003` | Major    | When `apiKey` is unset or empty, no authentication header **MUST** be sent.                                                                   |
| `GOFF-AUTH-004` | Minor    | The provider **SHOULD** allow arbitrary additional headers, for deployments behind gateways requiring their own authentication.               |

:::note Migrating from `X-API-Key`
The relay proxy accepts **both** `X-API-Key` and `Authorization: Bearer`, resolving `X-API-Key`
first. A provider can therefore switch to `Authorization: Bearer` unilaterally, with no
coordinated server release and no breaking change for users.
:::

---

## 16. Remote fallback — `GOFF-FALLBACK`

**Tier: In-process**

When local evaluation fails in a way that suggests the provider — rather than the flag — is at
fault, the relay proxy is authoritative and reachable. Falling back to it converts a local
failure into a correct answer.

| ID                  | Sev      | Requirement                                                                                                                                                                 |
| ------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GOFF-FALLBACK-001` | Major    | When in-process evaluation returns raw engine code `PARSE_ERROR` or `GENERAL`, the provider **MUST** retry the evaluation remotely via OFREP and return the remote result.   |
| `GOFF-FALLBACK-002` | Major    | The trigger **MUST** be evaluated against the **raw engine error code**, before mapping to the SDK's error enumeration.                                                      |
| `GOFF-FALLBACK-003` | Major    | `FLAG_CONFIG` **MUST NOT** trigger a fallback. It is a deterministic misconfiguration the relay proxy would reproduce identically.                                           |
| `GOFF-FALLBACK-004` | Major    | The fallback **MUST** be attempted on every qualifying occurrence.                                                                                                          |
| `GOFF-FALLBACK-005` | Major    | If the remote call also fails, the provider **MUST** return the **original in-process error**, and **MUST** log the remote failure. The in-process error is the root cause.  |
| `GOFF-FALLBACK-006` | Major    | A fallback result **MUST NOT** emit a feature event. The relay proxy has already recorded it.                                                                               |
| `GOFF-FALLBACK-007` | Major    | A fallback result **MUST** carry flag metadata `gofeatureflag_evaluated_remotely: true`.                                                                                     |
| `GOFF-FALLBACK-008` | Major    | Each fallback **MUST** be logged at warning level.                                                                                                                          |
| `GOFF-FALLBACK-009` | Major    | Authentication and timeout **MUST** apply to the fallback request identically to a normal remote evaluation.                                                                |

Where a fallback follows a WebAssembly trap, the ordering is: catch the trap, discard and
rebuild the instance (`GOFF-WASM-008`), then fall back. The provider **MUST NOT** retry locally
on the fresh instance first.

:::warning Operational consequence
Because `GOFF-FALLBACK-004` mandates fallback on every occurrence, a persistently malformed flag
turns every evaluation of it into a network round trip — an in-process provider silently
degrading to worse-than-remote latency. `GOFF-FALLBACK-007` and `GOFF-FALLBACK-008` exist so that
this condition is diagnosable rather than invisible.
:::

---

## 17. Remote cache (optional)

**Tier: Optional.** Supersedes the former Provider Cache specification.

A provider **MAY** cache remote evaluation results locally. If it does, the following apply; if
it does not, they are **N/A** and no feature events are produced in remote mode.

| ID               | Sev   | Requirement                                                                                                     |
| ---------------- | ----- | ----------------------------------------------------------------------------------------------------------------- |
| `GOFF-CACHE-001` | Major | The cache **MUST** be an LRU with configurable maximum size and TTL, defaulting to `10000` entries and `60 s`.   |
| `GOFF-CACHE-002` | Major | A TTL of `-1` **MUST** mean entries never expire.                                                               |
| `GOFF-CACHE-003` | Major | When present, the cache **MUST** be enabled by default and **MUST** be possible to disable by configuration.     |
| `GOFF-CACHE-004` | Major | The cache key **MUST** combine the flag key and the evaluation context.                                          |
| `GOFF-CACHE-005` | Major | Evaluation **MUST** consult the cache before calling the relay proxy, and **MUST** populate it on a miss.        |
| `GOFF-CACHE-006` | Major | Expired entries **MUST NOT** be served, and the oldest entry **MUST** be evicted when the maximum size is reached. |
| `GOFF-CACHE-007` | Major | Cache hits **MUST** produce feature events with `source: PROVIDER_CACHE` — they are invisible to the relay proxy, which never saw the request. |

---

## Appendix A — Endpoint reference

### A.1 Normative

| Method | Path                                     | Purpose                        |
| ------ | ---------------------------------------- | ------------------------------ |
| `POST` | `/ofrep/v1/evaluate/flags/{flagKey}`     | Remote single-flag evaluation  |
| `POST` | `/ofrep/v1/evaluate/flags`               | Remote bulk evaluation         |
| `POST` | `/v1/flag/configuration`                 | In-process configuration fetch |
| `POST` | `/v1/data/collector`                     | Evaluation and tracking events |

### A.2 Superseded

These remain functional but **MUST NOT** be used by a conformant provider.

| Method | Path                        | Replaced by                            |
| ------ | --------------------------- | -------------------------------------- |
| `POST` | `/v1/feature/{flagKey}/eval`| `/ofrep/v1/evaluate/flags/{flagKey}`   |
| `POST` | `/v1/allflags`              | `/ofrep/v1/evaluate/flags`             |
| `GET`  | `/ws/v1/flag/change`        | `/stream/v1/ws/flag/change` *(formally deprecated; the relay proxy returns RFC 8594 `Deprecation` headers)* |

---

## Appendix B — Conformance fixtures

These fixtures are canonical. They live in the
[go-feature-flag repository](https://github.com/thomaspoignant/go-feature-flag) and providers
**MUST** test against them unmodified. A provider whose local copy has diverged from canon is
non-conformant regardless of whether its own tests pass.

### B.1 Canonical evaluation context

```json
{
  "targetingKey": "d45e303a-38c2-11ed-a261-0242ac120002",
  "email": "john.doe@gofeatureflag.org",
  "firstname": "john",
  "lastname": "doe",
  "anonymous": false,
  "professional": true,
  "rate": 3.14,
  "age": 30,
  "company_info": { "name": "my_company", "size": 120 },
  "labels": ["pro", "beta"]
}
```

### B.2 In-process expected results

Source: `openfeature/providers/python-provider/tests/mock_responses/config/valid-all-types.json`,
evaluated with the context above. `evaluationContextEnrichment` is `{"env": "production"}`.

| Flag                        | Value             | Variant      | Reason            | Error code       |
| --------------------------- | ----------------- | ------------ | ----------------- | ---------------- |
| `bool_targeting_match`      | `true`            | `enabled`    | `TARGETING_MATCH` | —                |
| `string_key`                | `"CC0002"`        | `color1`     | `STATIC`          | —                |
| `double_key`                | `101.25`          | `medium`     | `TARGETING_MATCH` | —                |
| `integer_key`               | `101`             | `medium`     | `TARGETING_MATCH` | —                |
| `object_key`                | `{"test":"false"}`| `varB`       | `TARGETING_MATCH` | —                |
| `disabled_bool`             | *SDK default*     | `SdkDefault` | `DISABLED`        | —                |
| `DOES_NOT_EXIST`            | *SDK default*     | —            | `ERROR`           | `FLAG_NOT_FOUND` |
| `string_key` *(as boolean)* | *SDK default*     | —            | `ERROR`           | `TYPE_MISMATCH`  |

Notes that catch real bugs:

- `string_key` resolves via its default rule with no targeting match, so the reason is
  `STATIC`, not `TARGETING_MATCH`.
- `string_key` sets `trackEvents: false` and therefore **MUST NOT** produce a feature event
  (`GOFF-COLL-022`).
- A disabled flag returns the caller's default with variant `SdkDefault` — not `defaultSdk`,
  and not a null variant.

### B.3 Engine ABI vectors

Source: `cmd/wasm/testdata/`. Input:

```json
{
  "flagKey": "TEST",
  "flag": {
    "variations": { "enable": true, "disable": false },
    "targeting": [
      {
        "name": "targetingID rule",
        "query": "targetingKey eq \"random-key\"",
        "percentage": { "enable": 90, "disable": 10 }
      }
    ],
    "defaultRule": { "variation": "disable" },
    "metadata": { "description": "test flag", "type": "boolean" }
  },
  "evalContext": { "targetingKey": "random-key", "age": 42 },
  "flagContext": {
    "evaluationContextEnrichment": { "env": "production" },
    "defaultSdkValue": false
  }
}
```

Output:

```json
{
  "trackEvents": true,
  "variationType": "enable",
  "failed": false,
  "version": "",
  "reason": "TARGETING_MATCH_SPLIT",
  "errorCode": "",
  "value": true,
  "cacheable": true,
  "metadata": {
    "description": "test flag",
    "evaluatedRuleName": "targetingID rule",
    "type": "boolean"
  }
}
```

Two further vectors are canonical: an empty targeting key against a bucketing flag yields
`errorCode: "TARGETING_KEY_MISSING"` with `variationType: "SdkDefault"`, and a malformed input
yields `errorCode: "PARSE_ERROR"` with `value: null`. Both **MUST** trigger the behaviour of
[§16](#16-remote-fallback--goff-fallback) where applicable.

### B.4 Remote integration fixture

`openfeature/provider_tests/flags.yaml` drives the cross-language integration suites. Its
variations are named `Default`, `"False"` and `"True"`; the canonical context above matches the
targeting rule on every flag, so remote evaluation returns variant `True`. Each flag carries
metadata `description` and `pr_link`, which **MUST** be surfaced as flag metadata unmodified.

---

## Appendix C — Conformance checklist and report format

### C.1 Verdicts

Each requirement is reported as **PASS**, **FAIL** or **N/A**. `N/A` is reserved for
requirements outside the provider's declared tiers, or capabilities its SDK does not offer. A
requirement that applies but cannot be verified is **FAIL**, not `N/A`.

### C.2 Report format

A conformance report **MUST** open with a single verdict line, then list findings grouped by
severity, most severe first. Every finding cites its requirement identifier and the source
location that evidences it.

```text
GO Feature Flag Provider Specification 1.0 — <provider> <version>
VERDICT: NON-CONFORMANT — 2 Critical, 5 Major, 3 Minor (118 PASS, 10 FAIL, 9 N/A)

CRITICAL
  GOFF-IP-007   304 without ETag wipes the flag map        src/evaluator/inprocess.ts:142
  GOFF-CTX-006  Enrich hook replaces the gofeatureflag map src/hook/enrich.ts:18

MAJOR
  ...
```

### C.3 Auditing method

The procedure for producing such a report — which files to read, in what order, and how to
decide `N/A` — is maintained alongside the source rather than here, so it can evolve without a
specification revision.

One rule is normative because it was learned the hard way: **source is authoritative over
documentation.** During the audits that produced this specification, every provider's own
README, doc comments or agent instructions were wrong about that provider's behaviour, and in
several cases documented a default that no code path ever assigned. A contradiction between a
provider's documentation and its code **MUST** be reported as a finding in its own right.

---

## Appendix D — Known gaps

*This appendix is populated in a follow-up revision, once the corresponding issues have been
filed against each provider repository. Each row will cite its tracking issue.*
