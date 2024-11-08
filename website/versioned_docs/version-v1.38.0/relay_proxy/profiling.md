---
sidebar_position: 81
title: Profiling
description: Profiling of the relay proxy.
---

## Profiling

The **relay proxy** is able to expose profiling information.  
This is useful to understand the performance of the service and solve potential issues.

The information are exposed on the `/debug/pprof` endpoint, and we are using the default `net/http/pprof` package
to expose the information.

:::warning
By default the profiling endpoints are disabled.  
You have to run the relay proxy in debug mode if you want to enable them.
:::

List of endpoints exposed is available http://localhost:1031/debug/pprof/

### Enable profiling

In your relay proxy configuration file you need to set the `debug` field to `true`.

```yaml {5}
retriever:
  kind: file
  path: /goff/flags.yaml # Location of your feature flag files
#  ...
debug: true
```