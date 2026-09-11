---
title: Provider Cache Specification
---

# Specification Document for OpenFeature Providers Cache

|                      |                 |
| -------------------- | --------------- |
| **Creation Date**    | 06/04/2023      |
| **Last Update Date** | 03/08/2026      |
| **Authors**          | Thomas Poignant |

:::info This document has moved
The provider cache requirements are now part of the
[**GO Feature Flag Provider Specification**](/specification/openfeature-provider), in the
section _Remote cache (optional)_.

They were merged so that a single document defines the whole provider contract. Keeping the
cache policy separate is how it came to be implemented by only one provider while the others
quietly dropped it.
:::

## What changed

The requirements themselves are substantially unchanged — an LRU cache with configurable size
and TTL, enabled by default, keyed by flag and evaluation context, collecting usage for cache
hits only. Three points were clarified when they were merged:

- The cache is now explicitly **optional**. A provider that does not implement one is
  conformant; a provider that does must follow the requirements in full.
- Default size and TTL are now specified (`10000` entries, `60 s`), rather than only being
  required to be configurable.
- The reason cache hits are collected but remote evaluations are not is now stated as a
  general rule about event attribution, alongside the `source` field that records it.

For the current text, see
[**the provider specification**](/specification/openfeature-provider).
