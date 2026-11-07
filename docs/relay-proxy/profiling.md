# 📈 Profiling

The **relay proxy** is able to expose profiling information.<br /><!-- -->This is useful to understand the performance of the service and solve potential issues.

The information are exposed on the `/debug/pprof` endpoint, and we are using the default `net/http/pprof` package to expose the information.

warning

By default the profiling endpoints are disabled, and must be enabled in the configuration file.

List of endpoints exposed is available <http://localhost:1031/debug/pprof/>

### Enable profiling[​](#enable-profiling "Direct link to Enable profiling")

In your relay proxy configuration file you need to set the `enablePprof` field to `true`.

goff-proxy.yaml

```
retriever:

  kind: file

  path: /goff/flags.yaml

#  ...

enablePprof: true
```

note

The `debug` field also enables profiling, but it is deprecated.

goff-proxy.yaml

```
#...

debug: true
```
