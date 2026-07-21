# GO Feature Flag evaluation WASM module

This module compiles the GO Feature Flag evaluation logic (`modules/core`) to
WebAssembly with TinyGo. The resulting binaries are published to
[go-feature-flag/wasm-releases](https://github.com/go-feature-flag/wasm-releases)
and embedded by the OpenFeature in-process providers (Python, Java, .NET,
JavaScript) to evaluate flags without calling the relay proxy.

## Building

```bash
make build-wasi   # out/bin/gofeatureflag-evaluation.wasi (WASI hosts: wasmtime, chicory, ...)
make build-wasm   # out/bin/gofeatureflag-evaluation.wasm (browser/JS hosts)
```

Both targets use TinyGo with `-scheduler=none` and the custom target files in
`targets/`, which raise the shadow stack from the 64KB wasm-ld default to
1MB (`-z stack-size=1048576`). Do not remove that flag: with a 64KB stack,
realistic flag configurations (a targeting query with ~30 nested parentheses,
an `in` list with 154+ items — a ~200-integer allow-list caused the
production traps of issue
[#5651](https://github.com/thomaspoignant/go-feature-flag/issues/5651) —
or a ~340-condition `and`/`or` chain) overflow the stack and trap the
instance.
The build asserts the resulting binary carries that stack size
(`.github/ci-scripts/verify-wasm-stack.py`), so a TinyGo upgrade that stops
honoring the target-JSON `ldflags` fails the build instead of silently
shipping a 64KB binary again.

## ABI

Exports (besides the TinyGo runtime exports `calloc`/`realloc`/`_start`):

| Export | Signature | Purpose |
|---|---|---|
| `memory` | — | linear memory used for data exchange |
| `malloc` | `(size: i32) -> i32` | allocate a host-owned input buffer |
| `free` | `(ptr: i32)` | release a buffer previously returned by `malloc` |
| `evaluate` | `(ptr: i32, len: i32) -> i64` | evaluate a flag |

Call sequence for one evaluation:

1. Serialize the input (see `EvaluateInput` in `evaluate_input.go`) to UTF-8
   JSON.
2. `ptr = malloc(len + 1)`, then write the bytes plus a NUL terminator at
   `ptr`.
3. `result = evaluate(ptr, len)` (`len` excludes the NUL terminator).
4. Unpack `outputPtr = result >> 32` and `outputLen = result & 0xFFFFFFFF`,
   then read `outputLen` bytes at `outputPtr` and JSON-parse them
   (`model.VariationResult`). A `result` of `0` means no output was produced.
5. `free(ptr)`.

### Rules hosts MUST follow

- **Read the output before any further call into the module.** The output
  buffer belongs to the module's garbage collector; it is kept alive only
  until the next call into the instance (`helpers.lastOutput`).
- **One instance = one thread.** The module is built with `-scheduler=none`
  and is not reentrant. Never run two calls concurrently on the same
  instance; use an instance pool for parallelism.
- **A trapped instance is permanently poisoned — discard it.** A WASM trap
  (stack overflow, unrecoverable panic, out-of-memory) does not unwind the
  module's shadow stack pointer. After any trap, every later call is likely
  to fault inside `malloc` at a wrapped `0xffffXXXX` address. Hosts must
  drop the instance (including its store/memory) and instantiate a fresh
  one. Do not call `free` on a trapped instance.

## Built-in safeguards

`evaluate` never intentionally traps on bad input:

- Input JSON nested deeper than 128 levels returns a structured `PARSE_ERROR`
  result instead of overflowing the parser stack.
- Targeting queries have a per-format nesting budget: 64 nested
  brackets/parentheses for nikunjy expressions, 256 for JSONLogic documents
  (which spend ~5 bracket levels per logical operator). Exceeding it returns
  a structured `PARSE_ERROR`.
- nikunjy `[...]` lists are capped at 1,000 items per list; larger lists
  return a structured `PARSE_ERROR`. List parsing is right-recursive (one
  parser stack frame of ~356 bytes per item), so item count — not bracket
  nesting — drives stack use for flat lists: measured first-trap is 154 items
  on a 64KB stack and 2,947 on the 1MB stack, identical for int, double and
  string lists. JSONLogic arrays are exempt (decoded iteratively). Very
  large allow-lists should be split into `or`-joined `in` chunks or moved to
  JSONLogic.
- nikunjy `and`/`or` chains are capped at 1,000 conditions per query; larger
  chains return a structured `PARSE_ERROR`. Logical expressions are binary
  and recursive, so a flat bracket-less chain consumes parser stack per
  operator while being invisible to the nesting and list guards: measured
  first-trap is 341 conditions on a 64KB stack and 3,266 on the 1MB stack.
  JSONLogic documents are exempt (their operands live in iteratively-decoded
  arrays).
- Any Go panic during evaluation is recovered and returned as a `GENERAL`
  error result.

**Behavior change vs binaries <= 0.2.3:** inputs or queries beyond these
limits previously either evaluated by accident or trapped (permanently
poisoning the instance); they now deterministically return a structured
error, and the host falls back to the default value. The Python host applies
its own pre-flight nesting guard (also 128) before calling the module; the
two guards measure nesting slightly differently (Python object depth vs
serialized bracket depth) and are each intentionally conservative — they are
not guaranteed to reject the exact same set of inputs.

These guards cover every known overflow trigger (input nesting, query
nesting, `in`-list breadth, `and`/`or` chain length), but recursion inside
the module cannot be enumerated exhaustively, and older binaries in the
field carry none of these guards. The trap-handling rule above therefore
remains mandatory for hosts.
