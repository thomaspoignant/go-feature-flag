---
name: audit-provider
description: Audit a GO Feature Flag OpenFeature provider against the conformance specification and emit a PASS/FAIL/N-A report. Use when asked to audit, validate, or check conformance of a GO Feature Flag provider in any language.
---

# Audit a GO Feature Flag provider

Produce a conformance report for one provider against
[the specification](https://gofeatureflag.org/specification/openfeature-provider)
(`website/src/pages/specification/openfeature-provider.md` in this repo).

The specification is the source of requirements, verdict rules and report format. This skill
is only the *method*. If the two disagree, the specification wins.

Read the whole specification before auditing anything. In particular, Appendix C.1 defines
`PASS`/`FAIL`/`N/A` and the rules for **vacuous** and **accidental** satisfaction; do not
invent your own.

## Step 1 — Establish identity and tiers

Before reading any implementation code, determine:

| Question | Decides |
| --- | --- |
| Which language, and where is the source? | Where to read |
| Which evaluation modes does it support? | `Remote` / `In-process` tiers |
| Does it embed the engine via WebAssembly, or link it directly? | `WASM` tier, or N/A for all of §10 |
| Does the language's OpenFeature SDK have a Tracking API? | §14 applicability |
| Does it cache remote evaluations? | §17 applicability, and `GOFF-COLL-023` |

State the tier set in the report header. Every later `N/A` must trace back to it. A Go
provider is `N/A` for §10 because it links the engine directly — that is conformant, not a gap.

## Step 2 — Locate everything you will read

**Three sources, not one.** Missing the second and third is the most common way an audit
reaches wrong conclusions.

**a. The provider itself.** In this repository:

- `openfeature/providers/python-provider/gofeatureflag_python_provider/`
- `openfeature/providers/kotlin-provider/` *(client-side; out of scope for spec 1.0)*

External, via `gh api "repos/<repo>/contents/<path>" --jq '.content' | base64 -d`:

| Language | Repo | Base path |
| --- | --- | --- |
| .NET | `open-feature/dotnet-sdk-contrib` | `src/OpenFeature.Providers.GOFeatureFlag/` |
| Java | `open-feature/java-sdk-contrib` | `providers/go-feature-flag/` |
| JavaScript | `open-feature/js-sdk-contrib` | `libs/providers/go-feature-flag/` |
| Go | `open-feature/go-sdk-contrib` | `providers/go-feature-flag/` |

List files first with `git/trees/<branch>?recursive=1`. Quote the URL — `?` is a zsh glob.

**b. Delegated dependencies.** Identify, from the manifest, anything the provider defers
behaviour to — most often a generic OFREP client — and **locate its pinned source on disk
before you start**. Per §1.7 the provider owns that behaviour, so a defect there is a finding
against the provider. Where the pinned source lives:

| Language | Typically |
| --- | --- |
| Python | `~/.cache/uv/archive-v0/*/`, or the active virtualenv's `site-packages/` |
| JavaScript | `node_modules/`, or the sibling `libs/shared/` in a monorepo |
| Java | `~/.m2/repository/`, or fetch the source from the dependency's own repo |
| .NET | `~/.nuget/packages/<pkg>/<version>/lib/` |
| Go | `$(go env GOMODCACHE)/<module>@<version>/` |

**c. The language's OpenFeature SDK.** Needed to resolve §1.5 deferrals — the canonical
object type, the `ErrorCode` enum, how provider status transitions are derived from thrown
errors. Same locations as (b).

## Step 3 — Read in this order

Each stage establishes facts the next depends on.

1. **Options / configuration** — canonical names, defaults, which options exist. Gives §3.
2. **Provider entry point** — metadata, lifecycle ordering, hook registration, `track()`.
   Gives §2, §4, §12, §14.
3. **Evaluators** — type checks, reason handling, error mapping, flag lookup. Gives §6, §11, §9.3.
4. **HTTP / API layer** — endpoints, auth, ETag, 304 and status mapping, URL construction.
   Gives §9.1, §15, `GOFF-IP-003`. **If remote mode is delegated, §8 comes from the dependency
   located in Step 2b, not from the provider.**
5. **Engine host** — WASM ABI or direct linkage; pooling, trap handling. Gives §10.
6. **Hooks** — enrichment nesting, collector gating. Gives §7.1, §12.
7. **Event publisher and event models** — wire field names, flush, buffer policy. Gives §13.
8. **Tests and fixtures** — see Step 5.

## Step 4 — Run the targeted checks

Run these deliberately. They are not a summary of the specification; they are the specific
places where every audit so far has found real defects, and several are invisible unless you
go looking.

| Check | Why |
| --- | --- |
| Trace a `304` carrying **no** `ETag` header, line by line | Wiped the flag map in three of five providers; one also killed polling permanently |
| Trace a `200` with a valid but **empty** body | Wipes flags *and* pins the ETag, so the empty state never recovers |
| Check whether the 304 path is *structurally* distinct from a 200 | `GOFF-IP-007` requires a sentinel or distinct type, not an empty response object |
| Grep the failed-evaluation boolean's serialised name | One provider sends `defaultValue`; the server reads `default` and records failures as successes |
| Check whether the enrichment hook **merges** or **replaces** `gofeatureflag` | Replacing destroys caller-supplied `flagList` and `currentDateTime` |
| Check whether a trapped engine instance is returned to a pool | Every audited provider reused poisoned instances |
| Grep every declared option for a read site | Two providers had options that do nothing |
| Check whether a lock is held across the collector HTTP call | Couples flag-evaluation latency to collector availability |
| Check the timeout actually applied on each transport | One provider silently used a dependency's 5 s default for remote and its own 10 s elsewhere |
| Diff the provider's fixtures against Appendix B canon | They were byte-identical once; nothing enforces it |

WASM tier only: check the length argument passed to `evaluate` is the UTF-8 **byte** length.
This is unfalsifiable in languages whose strings are already byte-oriented — record it as PASS
with that reasoning rather than skipping it silently.

## Step 5 — Read the tests as evidence, not as proof

Tests are the fastest route to a provider's *intended* behaviour, and a rich source of two
finding classes the requirement list cannot express on its own:

- **Tests that assert non-conformant behaviour.** After establishing each `FAIL`, grep the
  suite for an assertion encoding it. A green suite pinning a defect means the behaviour was
  deliberate and remediation must change the test too. Report the test alongside the finding.
- **Tests that are commented out or skipped.** These mark behaviour the maintainers know is
  wrong or unstable. Note them where they touch a requirement.

Never treat a passing test as evidence that a requirement is met — it may assert the wrong
thing. Cite the implementation.

## Step 6 — Code over docs

**Never accept documentation as evidence of behaviour.** Every finding cites a source location
for *code*, not for a comment, README, or docstring.

During the audits that produced this specification, every provider's documentation was wrong
about that provider:

- a JSDoc `@default 120000` for a value no code path assigns, leaving the feature off by default
- an option declared in the type, documented in the README, asserted in tests, never read
- a README stating a hook is registered in one mode, contradicting code that registers it in both
- a Javadoc default of `1000 ms` against a constant of `60000 ms`
- an options comment claiming the default mode is `REMOTE` where the code says `INPROCESS`
- agent instructions describing an architecture removed several releases ago

Report contradictions per Appendix C.2. This class matters disproportionately for a repository
that other agents will read: an agent following stale instructions implements against a
superseded API.

## Step 7 — Assign verdicts and report

Verdict rules are in Appendix C.1 — including vacuous and accidental satisfaction, which you
must apply rather than improvise. Report format is Appendix C.2, including the per-requirement
table and the two extra finding classes.

Cover **every** requirement identifier. Do not sample.

Do not soften a Critical because it is unlikely, and do not inflate a Minor because it is easy
to fix. Severity is defined by consequence, and it is what the maintainer triages on.

## Scope

Audit **one** provider per run. Comparative claims ("better than the Java one") belong in
Appendix D of the specification, not in a single provider's report.
