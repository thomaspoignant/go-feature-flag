---
title: Implementing a GO Feature Flag Provider
---

# Implementing a GO Feature Flag Provider

|                          |                                                    |
| ------------------------ | -------------------------------------------------- |
| **Companion to**         | [Provider Specification 1.0](/specification/openfeature-provider) |
| **Creation Date**        | 03/08/2026                                          |
| **Last Update Date**     | 03/08/2026                                          |
| **Authors**              | Thomas Poignant                                     |

## Who this is for

You are writing a GO Feature Flag provider for a language that does not have one yet.

The [specification](/specification/openfeature-provider) tells you every rule your provider
must satisfy. It does not tell you what to build, in what order, or what the payloads look
like end to end — an auditor already has the code in front of it, so the specification only
describes constraints. This guide supplies the shape.

**The specification is normative; this guide is not.** Where they appear to disagree, the
specification wins. This guide links to requirement identifiers rather than restating rules,
so the two cannot drift.

:::tip
Build **remote mode first**. It is a fraction of the work, it exercises the whole
configuration, hook and telemetry stack, and it gives you something releasable before you take
on the evaluation engine.
:::

---

## 1. Decisions to make before writing code

| Decision | Guidance |
| --- | --- |
| Remote only, or remote + in-process? | In-process needs a WebAssembly runtime for your language. If none exists, ship remote-only — it is a legitimate conformance tier, not a lesser provider. |
| Which WASM runtime? | Existing providers use Wasmtime (Python, .NET), Chicory (Java) and the platform's built-in `WebAssembly` (JavaScript). Any runtime supporting WASI works. |
| Write your own OFREP client, or reuse one? | Reuse if your ecosystem has one — but read [§1.7](/specification/openfeature-provider#17-delegated-behaviour) first. You remain accountable for its behaviour, and delegated OFREP clients have been the source of real conformance failures. |

---

## 2. Components

Every existing provider converged on the same decomposition. You are not obliged to copy it,
but it maps cleanly onto the specification's sections, which makes conformance easier to
demonstrate.

```text
Provider                     ← OpenFeature entry point: metadata, lifecycle, hooks, track()
├── Options                  ← configuration, validation, defaults
├── Evaluator (interface)
│   ├── RemoteEvaluator      ← OFREP calls
│   └── InProcessEvaluator   ← local config + engine + polling
│       └── EngineHost       ← WASM instance pool, or a direct binding
├── ApiClient                ← flag configuration + data collector HTTP
├── EventPublisher           ← event buffer, flush loop
└── Hooks
    ├── EnrichEvaluationContextHook   ← before:  injects exporterMetadata
    └── DataCollectorHook             ← after/error: emits feature events
```

The **evaluator interface** is the load-bearing choice. Both modes must satisfy the same
five resolvers plus a trackability query, which is what lets the provider shell stay
mode-agnostic and keeps evaluation results identical between modes.

---

## 3. Build order

### Stage 1 — Options and the provider shell

Implement options with the names and defaults in
[§3](/specification/openfeature-provider#3-configuration--goff-cfg). Validate `endpoint` at
construction, before any I/O.

Two traps worth avoiding from the start, both of which are live defects in shipped providers:

- Copy the caller's options before normalising them (`GOFF-CFG-004`). Mutating a map the
  caller still holds causes cross-talk between two providers built from one config object.
- Do not read environment variables for the endpoint or credentials (`GOFF-CFG-005`).

### Stage 2 — Remote evaluation

Implement the five resolvers against OFREP. The type rules in
[§6](/specification/openfeature-provider#6-evaluation-api--goff-eval) are where most providers
get this wrong — in particular a boolean must not satisfy a numeric resolver, and the float
resolver must accept an integral JSON number.

Pass `reason` through as an **opaque string** (`GOFF-EVAL-007`). Parsing it into an enum is the
single most common defect found in audits: the engine emits `TARGETING_MATCH_SPLIT`, `SPLIT`
and `OFFLINE`, which most SDK `Reason` enums do not contain, so a lookup throws and turns a
correct evaluation into an error.

### Stage 3 — Hooks and telemetry

Add both hooks, in the order `[Enrich, DataCollector]`
([§12](/specification/openfeature-provider#12-hooks--goff-hook)), then the event publisher
([§13](/specification/openfeature-provider#13-data-collection--goff-coll)).

The enrichment hook must **merge** into the `gofeatureflag` namespace rather than replace it
(`GOFF-CTX-006`) — that namespace also carries caller-supplied `flagList` and `currentDateTime`.

### Stage 4 — In-process evaluation

Fetch and poll the flag configuration
([§9](/specification/openfeature-provider#9-in-process-evaluation--goff-ip)), then wire the
engine ([§10](/specification/openfeature-provider#10-webassembly-abi--goff-wasm)).

Get the 304 handling right *first*, before anything else in this stage. `GOFF-IP-007` requires
the "not modified" path to be structurally incapable of carrying a body — return a distinct
sentinel or type from your HTTP layer rather than an empty response object. Three of five
existing providers wipe their entire flag map on a 304 that omits an `ETag`, and every one of
them would have been prevented by that single structural choice.

### Stage 5 — Remote fallback

[§16](/specification/openfeature-provider#16-remote-fallback--goff-fallback). Straightforward
once both modes exist, since it is just routing an in-process failure to the remote path.

---

## 4. Wire payloads

### 4.1 Flag configuration

```http
POST {endpoint}/v1/flag/configuration
Content-Type: application/json
Authorization: Bearer {apiKey}
If-None-Match: "6c8a1f..."
```

```json
{ "flags": [] }
```

An empty array means "all flags". Response:

```json
{
  "flags": {
    "bool_targeting_match": {
      "variations": { "disabled": false, "enabled": true },
      "targeting": [
        { "query": "email eq \"john.doe@gofeatureflag.org\"", "variation": "enabled" }
      ],
      "defaultRule": { "percentage": { "enabled": 0, "disabled": 100 } },
      "metadata": { "description": "this is a test flag", "defaultValue": false }
    },
    "string_key": {
      "trackEvents": false,
      "variations": { "color1": "CC0002", "color2": "CC0001", "color3": "CC0000" },
      "defaultRule": { "percentage": { "color3": 0, "color2": 0, "color1": 100 } }
    }
  },
  "evaluationContextEnrichment": { "env": "production" }
}
```

:::danger Do not model the flag object
Store and forward each value under `flags` **exactly as received** (`GOFF-IP-016`). Do not
deserialise it into typed classes. The engine owns that schema and will extend it; a typed
model silently drops the additions and produces wrong values with no error.

The only field you may read is `trackEvents`, for `GOFF-COLL-020`.
:::

Response headers you must handle: `ETag`, stored and echoed **verbatim including quotes**
(`GOFF-IP-005`), and `Last-Modified`.

### 4.2 Remote evaluation

```http
POST {endpoint}/ofrep/v1/evaluate/flags/{flagKey}
```

```json
{
  "context": {
    "targetingKey": "d45e303a-38c2-11ed-a261-0242ac120002",
    "email": "john.doe@gofeatureflag.org",
    "anonymous": false
  }
}
```

```json
{
  "key": "bool_targeting_match",
  "value": true,
  "reason": "TARGETING_MATCH",
  "variant": "True",
  "metadata": { "test": "test1", "test2": false, "test3": 123.3 }
}
```

`variant` here is the same concept the engine calls `variationType` in-process. Map both to
your SDK's variant field.

### 4.3 Data collector

```http
POST {dataCollectorBaseURL}/v1/data/collector
```

```json
{
  "meta": { "provider": "rust", "openfeature": true, "myAppVersion": "1.4.0" },
  "events": [
    {
      "kind": "feature",
      "contextKind": "user",
      "userKey": "d45e303a-38c2-11ed-a261-0242ac120002",
      "creationDate": 1785772800,
      "key": "bool_targeting_match",
      "variation": "enabled",
      "value": true,
      "default": false,
      "version": "1.0.0",
      "source": "INPROCESS"
    },
    {
      "kind": "tracking",
      "contextKind": "user",
      "userKey": "d45e303a-38c2-11ed-a261-0242ac120002",
      "creationDate": 1785772800,
      "key": "checkout_completed",
      "evaluationContext": { "targetingKey": "d45e303a-38c2-11ed-a261-0242ac120002" },
      "trackingEventDetails": { "value": 29.99, "currency": "EUR" }
    }
  ]
}
```

Three fields are routinely got wrong: the envelope key is `meta`, not `metadata`; the failed
flag is `default`, not `defaultValue` — anything else is silently discarded by the relay proxy,
recording failures as successes; and `creationDate` is in **seconds**.

---

## 5. The evaluation engine

### Getting the binary

Releases live at [`go-feature-flag/wasm-releases`](https://github.com/go-feature-flag/wasm-releases),
as `evaluation/gofeatureflag-evaluation_{version}.wasi`. Existing providers vendor it as a git
submodule pinned to a commit, and copy the binary in at build time rather than committing it.

Pin the version in one machine-readable place (`GOFF-ENG-002`) — a version file, a build
property, whatever your ecosystem uses — and make it overridable at runtime so bundlers and
unusual packaging layouts stay usable (`GOFF-WASM-011`).

The version you pin is not arbitrary: it must be the one the
[specification names](/specification/openfeature-provider#16-evaluation-engine-version).
Conformance is defined relative to a specific engine, because two providers on different engine
builds can return different values for identical input.

### Calling it

The ABI is in [§10](/specification/openfeature-provider#10-webassembly-abi--goff-wasm). Input:

```json
{
  "flagKey": "bool_targeting_match",
  "flag": { "...the object from /v1/flag/configuration, untouched..." },
  "evalContext": { "targetingKey": "user-123", "email": "a@b.com" },
  "flagContext": {
    "defaultSdkValue": false,
    "evaluationContextEnrichment": { "env": "production" }
  }
}
```

Output:

```json
{
  "value": true,
  "variationType": "enabled",
  "reason": "TARGETING_MATCH",
  "errorCode": "",
  "errorDetails": "",
  "trackEvents": true,
  "cacheable": true,
  "version": "1.0.0",
  "failed": false,
  "metadata": { "description": "this is a test flag" }
}
```

An empty `errorCode` means success. Read `version` — it feeds `GOFF-COLL-009`, and most
existing providers forget to decode it.

Two things that have bitten every implementation:

- The length passed to `evaluate` is the **UTF-8 byte length**, not a string length. In
  languages with UTF-16 strings these differ, and the payload truncates on the first non-ASCII
  character (`GOFF-WASM-003`).
- A trapped instance has undefined memory and must be **discarded and rebuilt**, never returned
  to a pool (`GOFF-WASM-008`). Every provider audited got this wrong.

---

## 6. Lifecycle

Initialization, in order:

1. Validate options and build the API client.
2. Start the event publisher.
3. Initialize the evaluator — in-process: load the engine, fetch configuration **synchronously**,
   then start polling. Remote: nothing to do.
4. Register hooks, `[Enrich, DataCollector]`.

Initialization blocks until the provider can serve or fails (`GOFF-LIFE-001`). A `401`/`403`
must end in `PROVIDER_FATAL` (`GOFF-EVT-007`); anything else in `ERROR`, so the provider
recovers unattended when the relay proxy returns.

Shutdown reverses it: stop polling, flush buffered events, stop the publisher — in **both**
modes, whether or not data collection is enabled (`GOFF-LIFE-004`).

Make initialization safe to call twice (`GOFF-LIFE-002`). Cancel and join any existing poller,
and reset one-shot shutdown guards. Getting this wrong has produced duplicate pollers, a
process-killing panic, and a deadlocked shutdown in shipped providers.

---

## 7. Local development

Run a relay proxy against a flag file:

```yaml title="docker-compose.yml"
services:
  goff:
    image: 'thomaspoignant/go-feature-flag'
    ports:
      - '1031:1031'
    environment:
      - SERVER_PORT=1031
      - SERVER_MODE=http
      - POLLINGINTERVAL=1000
      - RETRIEVER_KIND=file
      - RETRIEVER_PATH=/config.goff.yaml
      - AUTHORIZEDKEYS_EVALUATION=apikey1
    volumes:
      - ./config.goff.yaml:/config.goff.yaml
```

Use the canonical flag file from
[Appendix B](/specification/openfeature-provider#appendix-b--conformance-fixtures) so your
results are comparable with every other provider. `AUTHORIZEDKEYS_EVALUATION` lets you exercise
both the authenticated and unauthenticated paths.

---

## 8. Testing

Use the canonical fixtures unmodified. A provider whose fixtures have drifted from canon is
non-conformant even if its own tests pass.

Cover at minimum, in both modes:

- the four-case matrix per type — valid, disabled, wrong type, not found
- a `304` **with no `ETag` header** — the flag map must survive
- a `200` with an empty or null flag map — configuration must survive and the ETag must not advance
- an engine trap — the instance must be replaced, not reused
- a flag with `trackEvents: false` — no feature event
- a flag absent from configuration — still trackable

One caution learned from auditing: several existing providers have tests that **assert
non-conformant behaviour**, which pins the defect and raises the cost of fixing it. When a test
and the specification disagree, change the test.

---

## 9. Before you claim conformance

Run the audit. In this repository:

```console
/audit-provider
```

It produces a PASS/FAIL/N-A verdict for every requirement in the specification, grouped by
severity. Expect `GOFF-FALLBACK-*` to be your last section, and treat any Critical as a release
blocker — that severity is reserved for wrong flag values, silent data loss and unrecoverable
state.

Then add your provider to the [SDK list](/docs/sdk) and state which specification version you
target (`GOFF-ENG-003`).
