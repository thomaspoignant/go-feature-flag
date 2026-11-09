# Web provider playground & stream blocker

Two small **dev tools** for manually testing how the relay-proxy pushes flag changes to clients (SSE and WebSocket), and how
clients fall back to polling when those streams break.

> [!NOTE]
> These are local dev tools, not production code. They are excluded from SonarCloud (`**/testdata/**`) and from the
> Scorecard results.

## What is in this folder?

| Tool | File(s) | What it does |
|------|---------|--------------|
| **Playground** | `index.html`, `src/main.ts`, `vite.config.ts` | A browser page that runs the **real** [GO Feature Flag web provider](https://github.com/open-feature/js-sdk-contrib/tree/main/libs/providers/go-feature-flag-web) against a relay-proxy. It shows every provider event, the provider logs, the connection status, and live flag values. |
| **Stream blocker** | `stream-blocker.mjs` | A tiny proxy (Node.js, no dependencies) that sits in front of the relay-proxy. It forwards everything, but on demand it **cuts only the flag-change streams** (`/stream/*` for SSE, `/ws/*` for WebSocket). The REST endpoints keep working. |
| Local relay-proxy | `docker-compose.yml`, `relay-proxy/goff-proxy.yaml`, `relay-proxy/flags.goff.yaml` | A relay-proxy that reads `flags.goff.yaml` every second. Edit that file, and every connected client is notified. |

## How the pieces fit together

```
 ┌──────────────────────┐        ┌───────────────────────┐        ┌──────────────────────────┐
 │ Browser              │        │ stream blocker        │        │ relay-proxy              │
 │ playground  :5174    │ ─────► │ :1032                 │ ─────► │ :1033 (docker)           │
 │ (web provider)       │        │                       │        │ :1031 (go run)           │
 └──────────────────────┘        │ /v1/*, /ofrep/*  ✅    │        │                          │
                                 │ /stream/*, /ws/* ✅/❌ │        │ reads flags.goff.yaml    │
                                 └───────────────────────┘        └──────────────────────────┘
```

- To test **SSE / WebSocket only**, the playground can talk to the relay-proxy directly. You don't need the blocker.
- To test the **polling fallback**, put the blocker in the middle. It breaks the stream while polling (`/v1/*`, `/ofrep/*`)
  still works, so you can see the provider keep updating flags through polling.

## Prerequisites

- **Node.js 20+** (for the playground and the blocker)
- **Docker** (for the relay-proxy in option A)
- **Go** (only for option B: running the relay-proxy from this repository's source)

Run every command below from this folder:

```shell
cd cmd/relayproxy/testdata/web-provider-playground
```

---

## Step 1: Start a relay-proxy

Pick **one** option.

### Option A: Docker, using the latest published image

```shell
docker compose up -d
```

Check that it is up:

```shell
curl http://localhost:1033/health
# {"initialized":true}
```

The relay-proxy listens on **`http://localhost:1033`**.

### Option B: From source, to test your own relay-proxy changes

```shell
RETRIEVER_PATH=./relay-proxy/flags.goff.yaml go run ../.. --config relay-proxy/goff-proxy.yaml
```

- `go run ../..` builds and starts `cmd/relayproxy`.
- `RETRIEVER_PATH` overrides the path from `goff-proxy.yaml` (`/goff/...` only exists inside the docker container).

Check that it is up:

```shell
curl http://localhost:1031/health
# {"initialized":true}
```

The relay-proxy listens on **`http://localhost:1031`**.

> [!TIP]
> Whichever option you pick, **`relay-proxy/flags.goff.yaml` is live**. Save a change and the relay-proxy picks it up
> within a second, then notifies every connected SSE / WebSocket client.

---

## Step 2: Start the stream blocker (needed for the polling fallback only)

In a new terminal, point it at the relay-proxy from step 1:

```shell
TARGET=http://localhost:1033 npm run blocker   # option A (docker)
TARGET=http://localhost:1031 npm run blocker   # option B (go run)
```

You should see:

```
stream-blocker on :1032 -> http://localhost:1033
```

| Env var  | Default                 | Meaning                                  |
|----------|-------------------------|------------------------------------------|
| `TARGET` | `http://localhost:1031` | The relay-proxy to forward requests to.  |
| `PORT`   | `1032`                  | The port the blocker listens on.         |

You can control the blocker from the playground UI (step 3), or from the command line:

```shell
curl "http://localhost:1032/__blocker"          # read the state   -> {"blocked":false}
curl "http://localhost:1032/__blocker?block=1"  # cut SSE + WS     -> {"blocked":true}
curl "http://localhost:1032/__blocker?block=0"  # let them through -> {"blocked":false}
```

While blocked:

- Open SSE / WebSocket connections are **closed right away**.
- New `/stream/*` requests get an **HTTP 503**, and WebSocket upgrades are refused.
- Everything else (`/v1/*`, `/ofrep/*`, `/health`, ...) **still goes through**.

The terminal logs `streams BLOCKED` / `streams allowed` each time you switch.

---

## Step 3: Start the playground

In a new terminal:

```shell
npm install
npm run dev
```

Open **http://localhost:5174**.

### Which provider version runs?

By default, the playground uses the **published npm package** `@openfeature/go-feature-flag-web-provider`.

To test an **unreleased provider** (for example a branch of `open-feature/js-sdk-contrib`), point `PROVIDER_SRC` at your
local checkout:

```shell
PROVIDER_SRC=/path/to/js-sdk-contrib npm run dev
```

The page now loads the provider straight from `libs/providers/go-feature-flag-web/src` in that checkout, and edits to the
provider reload in the browser right away. The page title shows which provider is running: `(provider: npm package)` or
`(provider: source: /path/...)`.

### Tour of the page

| Area | What it is for |
|------|----------------|
| **Provider options** | Every option passed to `new GoFeatureFlagWebProvider(...)`. Set `endpoint` to the relay-proxy (`:1033` / `:1031`) or to the blocker (`:1032`). `mode` is `sse` or `ws`. The values are saved in `localStorage`. |
| **Connect / Disconnect** | **Connect** builds the provider and calls `OpenFeature.setProviderAndWait()`. **Disconnect** calls `OpenFeature.clearProviders()`, which runs the provider's `onClose`. |
| **Evaluation context** | `targetingKey` plus extra attributes as JSON. **setContext()** calls `OpenFeature.setContext()`, which runs the provider's `onContextChange`. |
| **Status** | **SDK status** is the OpenFeature client status (`READY`, `STALE`, `ERROR`...). **Last event** is the last provider event. **Change strategy** is the provider's connection state (`sse: connected`, `ws: error`, ...). |
| **Polling fallback** | The blocker status and the **Block SSE/WS** / **Allow SSE/WS** buttons. They only work when the blocker runs on `:1032`. |
| **Flags** | The flags to evaluate, with value, variant, reason, and error. Every provider event re-evaluates them, and a row **flashes** when its value changes. Add or remove flag keys as you like. |
| **Log** | Provider events plus everything the provider logs (`debug` / `info` / `warn` / `error`). Untick `debug` to hide the noisy lines. |

> [!TIP]
> The page exposes `window.provider` and `window.OpenFeature`, so you can poke at them from the browser devtools console.

---

## Guided scenarios

Each scenario lists **what to do** and **what you should see**.

### 1. Connect in SSE, then in WebSocket

1. Set `endpoint` to your relay-proxy (`http://localhost:1033` or `:1031`), set `mode` to `sse`, then click **Connect**.
   - ✅ The log shows `PROVIDER_READY`. SDK status is `READY`, change strategy is `sse: connected`, and the flags table shows values.
2. Click **Disconnect**, set `mode` to `ws`, then click **Connect**.
   - ✅ Same result, with `ws: connected`.

### 2. A flag changes

1. While connected, open `relay-proxy/flags.goff.yaml`. Change `my-string-flag`'s `defaultRule.variation` from `blue`
   to `red`, then save.
   - ✅ Within about a second, the log shows `PROVIDER_CONFIGURATION_CHANGED` with `flagsChanged: ["my-string-flag"]`, and the row flashes.

### 3. A flag is deleted

1. Remove the whole `my-number-flag` block from `flags.goff.yaml`, then save.
   - ✅ `PROVIDER_CONFIGURATION_CHANGED`, and the row shows `FLAG_NOT_FOUND` in the error column.

### 4. The evaluation context changes

1. In the context JSON, set `"email": "john@gofeatureflag.org"`, then click **setContext()**.
2. Change it to `"email": "john@example.com"`, then click **setContext()** again.
   - ✅ Each call shows `PROVIDER_RECONCILING`, then `PROVIDER_CONTEXT_CHANGED`. The values are evaluated again for
     the new context (`my-bool-flag` / `my-string-flag` depend on the email).

### 5. Full outage of the relay-proxy

1. While connected, run `docker compose stop` (or stop `go run` with Ctrl-C).
   - ✅ The status goes to `STALE`, then `ERROR`. The log shows retries with growing delays (`retryInitialDelay` × `retryDelayMultiplier`).
2. Start it again: `docker compose start` (or run `go run ...` again).
   - ✅ The provider reconnects and comes back to `READY`.

### 6. Polling fallback (needs the blocker)

1. Start the blocker (step 2). Set `endpoint` to **`http://localhost:1032`**, set `pollingIntervalMs` > 0 (for example
   `5000`), then click **Connect**.
   - ✅ `sse: connected` (or `ws: connected`), and the blocker badge says `allowed`.
2. Click **Block SSE/WS**.
   - ✅ The change strategy goes to `error`, and the SDK goes `STALE`. After `maxRetries` attempts, the log shows `start polling cycle`.
3. Edit a flag in `flags.goff.yaml`.
   - ✅ Within `pollingIntervalMs`, the row updates. Polling uses `/v1/*`, which the blocker still lets through.
4. Click **Allow SSE/WS**.
   - ✅ On the next poll tick, the stream reconnects: the change strategy is `connected`, and the log shows `stop polling cycle`.

### 7. API key

1. Add `authorizedKeys` to `relay-proxy/goff-proxy.yaml`, then restart the relay-proxy:
   ```yaml
   authorizedKeys:
     evaluation:
       - my-secret-key
   ```
2. Connect **without** `apiKey`.
   - ❌ Expect `401` errors in the log.
3. Set `apiKey` to `my-secret-key` and connect again.
   - ✅ `READY`. The SSE / WebSocket URL carries the key as `?apiKey=...`, because `EventSource` can't send headers.

---

## Using the stream blocker on its own

The blocker has nothing to do with the web provider. It's handy for testing the reconnect or fallback logic of **any**
client (another SDK, a server-side provider, `curl`...):

```shell
TARGET=http://localhost:1031 PORT=1032 node stream-blocker.mjs
# point your client at http://localhost:1032, then:
curl "http://localhost:1032/__blocker?block=1"   # break the streams
curl "http://localhost:1032/__blocker?block=0"   # restore them
```

How it works, in short:

- **HTTP** (including SSE) is forwarded with `http.request`. Headers are flushed right away, so `EventSource` fires
  `open` without waiting for the first event.
- **WebSocket upgrades** are tunnelled as raw TCP to the target. The blocker never parses WebSocket frames.
- Blocking closes the open stream sockets and refuses new ones. Nothing else is touched.

---

## Troubleshooting

| Symptom | Cause / fix |
|---------|-------------|
| `EADDRINUSE ... :1032` when starting the blocker | A blocker is already running. Stop it, or use `PORT=1042 ...`. Note that the playground buttons always talk to `:1032`. |
| `Port 5174 is in use` | Another playground runs already. Stop it, or run `npx vite --port 5175`. |
| Blocker badge says `blocker not running` | The blocker is not on `:1032`. Start it (step 2), then reload the page. |
| Change strategy says `n/a` | The npm provider in use has no SSE support yet. Use `PROVIDER_SRC=...` (step 3). |
| SSE stays `connecting` in Firefox, with a warning after 5s | Firefox fires `open` only when the first body bytes arrive. Older relay-proxies send nothing until the first change. Upgrade the relay-proxy, or go through the blocker, which flushes headers. |
| CORS errors on a remote relay-proxy | The relay-proxy must allow the `http://localhost:5174` origin. Locally, the default config allows everything. |
| `flags.goff.yaml` edits are not picked up (option B) | Check that `RETRIEVER_PATH=./relay-proxy/flags.goff.yaml` is set, and that you run from this folder. |

## Cleanup

```shell
docker compose down     # option A
# then Ctrl-C the playground, the blocker, and `go run` (option B)
```
