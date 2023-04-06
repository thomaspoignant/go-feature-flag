---
title: Provider Cache Specification
---

# Specification Document for OpenFeature Providers Cache

|                      |                 |
|----------------------|-----------------|
| **Creation Date**    | 06/04/2023      |
| **Last Update Date** | 06/04/2023      |
| **Authors**          | Thomas Poignant |

## Overview

OpenFeature Providers are used to call the GO Feature Flag the relay proxy to evaluate flags.
The purpose of this specification document is to outline the requirements for implementing the cache policy in the 
providers that is compatible with GO Feature Flag.

## Requirements

1. The provider should implement an LRU cache with a configurable size and TTL.
2. The cache size and TTL should be configurable by the user of the provider.
3. The cache should be activated by default, but it should be possible to deactivate it in the configuration.
4. The cache key should be a combination of the flag name and user key.
5. When a flag is evaluated, the provider should first check the cache for the result before querying the relay proxy.
6. When there is no entry in the cache, we should query the relay proxy and update the cache with this entry.
7. When we have reached the max time for a cache entry, it should not be possible to retrieve it.
8. When we have reached the max size of the cache we should remove the oldest entry.
9. We should collect all the flags usage and call the API `/v1/data/collector` for statistics reason.
10. The data collection should be activated by default, but it should be possible to deactivate it in the configuration.
11. We should not call the relay proxy API `/v1/data/collector` for every evaluation, but call it in bulk with a periodic threshold.
12. The threshold should be configurable by the user of the provider.
