# Building OpenFeature Server Providers for GO Feature Flag

> **This document has been replaced.**
>
> The current, normative specification lives at
> **<https://gofeatureflag.org/specification/openfeature-provider>**

## Why

This file described how to build a GO Feature Flag provider, generalised from the .NET
implementation. It has been superseded by a specification derived from auditing **all five**
server-side providers (Python, .NET, Java, JavaScript, Go) against the relay proxy and the
evaluation engine source.

That audit found this document to be inaccurate in several load-bearing ways. It is recorded
here so that anyone who implemented against it knows what to re-check:

| This document said                               | Actual behaviour                                              |
| ------------------------------------------------ | ------------------------------------------------------------- |
| Data collector envelope key `metadata`           | `meta`                                                        |
| Event `source: "PROVIDER"`                       | `INPROCESS`, or `PROVIDER_CACHE` for a remote cache hit        |
| Tracking event `kind: "custom"`                  | `tracking` — the relay proxy dispatches on this literal        |
| Flag rules use `percentageRollout`               | `percentage`                                                   |
| Enrichment hook flattens metadata into the context | Nests it under `gofeatureflag.exporterMetadata`              |
| Feature event `default` is the default *value*   | It is a *boolean* meaning "the evaluation failed"             |
| `flushIntervalMs` default `1000`                 | `60000`                                                        |
| Flag format at `internal/flag/internal_flag.go`  | Moved to `modules/core/flag/internal_flag.go`                  |

The new specification also covers ground this one did not: provider events and status
transitions, 304 and cache-validation semantics, evaluation-context normalisation, remote
fallback, conformance tiers, and a set of canonical fixtures with expected results.

## For implementers

Start with the specification linked above. It is written to be checked mechanically: every
requirement has a stable identifier, an RFC 2119 keyword and a severity, and Appendix B
contains the canonical fixtures your provider should test against.
