# Runtimes

Runtimes execute compiled wasm modules. Each is wrapped so a module can be run
with `bazel run`, hermetically.

## wazero

[wazero](https://wazero.io) is a WebAssembly 1.0/2.0 runtime written in pure Go
with zero platform dependencies — ideal as the first runtime because it needs no
system libraries and builds cleanly under rules_go.

### CLI: `wazero-run`

```sh
bazel run //runtimes/wazero/cmd/wazero-run -- /abs/path/to/module.wasm [args...]
```

Instantiates the module with WASI preview-1, wires up stdio, argv, and the host
environment, and propagates the module's exit code. Because `bazel run` changes
the working directory, pass an **absolute** module path when invoking the CLI
directly. For a turnkey experience, prefer the macro below.

### Macro: `wazero_run`

`runtimes/wazero/defs.bzl` provides a macro that bundles a wasm module and the
runner as runfiles, producing a location-independent `bazel run` target:

```python
load("//runtimes/wazero:defs.bzl", "wazero_run")

wazero_run(
    name = "run",
    module = ":hello_wasm",        # a go_cross_binary or .wasm file
    module_args = ["world"],        # args passed to the module
)
```

```sh
bazel run //examples/hello-go-wasm:run -- extra args here
```

Arguments after `--` are appended to `module_args`.

## External CLI runtimes

The four most popular standalone runtimes — **Wasmtime**, **Wasmer**,
**WasmEdge**, and **WAMR** — ship as prebuilt command-line tools rather than Go
libraries. Each is fetched hermetically as an `http_archive` (pinned by SHA-256
in `MODULE.bazel`; linux/amd64) and wrapped by a `<name>_run` macro that mirrors
`wazero_run`, so the same `bazel run` UX applies to all of them:

| Runtime  | Org / language          | Macro           | Repo        |
| -------- | ----------------------- | --------------- | ----------- |
| wazero   | Tetrate / Go            | `wazero_run`    | (in-tree)   |
| Wasmtime | Bytecode Alliance / Rust| `wasmtime_run`  | `@wasmtime` |
| Wasmer   | wasmerio / Rust         | `wasmer_run`    | `@wasmer`   |
| WasmEdge | CNCF / C++              | `wasmedge_run`  | `@wasmedge` |
| WAMR     | Bytecode Alliance / C   | `wamr_run`      | `@wamr`     |

The hello and Lua examples run unmodified on every one of them:

```sh
bazel run //examples/hello-go-wasm:run            # wazero
bazel run //examples/hello-go-wasm:run_wasmtime   # Wasmtime
bazel run //examples/hello-go-wasm:run_wasmer     # Wasmer
bazel run //examples/hello-go-wasm:run_wasmedge   # WasmEdge
bazel run //examples/hello-go-wasm:run_wamr       # WAMR iwasm

# With a host directory mounted (the Lua interpreter reads its script):
bazel run //interpreters/golua:run_fib_wasmtime
bazel run //interpreters/golua:run_fib_wasmer
bazel run //interpreters/golua:run_fib_wasmedge
bazel run //interpreters/golua:run_fib_wamr
```

### How the wrapper works: `cli_wasm_run`

`runtimes/cli_run.bzl` provides one shared rule, `cli_wasm_run`, that carries the
wasm module, the runtime CLI binary, and the CLI's runtime files (e.g.
WasmEdge's `libwasmedge.so`) as runfiles, then generates a location-independent
wrapper script. The per-runtime `defs.bzl` macros only supply the bits that
differ between runtimes:

- **run prefix** — `wasmtime`/`wasmer` use a `run` subcommand; `wasmedge`/`iwasm`
  take the module directly. `iwasm` additionally defaults to `--interp`
  (its prebuilt fast-JIT exhausts its code cache on larger modules) with an
  8 MB `--stack-size` (the 64 KB default overflows on Go's deep call stacks).
- **mount syntax** — every runtime spells directory mounts differently:
  `--dir HOST::GUEST` (Wasmtime), `--volume HOST:GUEST` (Wasmer),
  `--dir GUEST:HOST` (WasmEdge), `--map-dir=GUEST::HOST` (WAMR).
- **arg separator** — Wasmer needs a `--` between the module and its arguments.

Two portability wrinkles the wrapper handles:

- **Runfiles symlinks vs. WASI sandboxes.** Bazel materializes runfiles as
  symlinks back into the source tree. wazero follows them, but the strict WASI
  sandboxes refuse to follow a symlink that escapes a preopened directory. So
  for any mount, the wrapper first copies the directory into a dereferenced
  temp copy (`cp -RL`) and mounts that, cleaning up on exit.
- **WasmEdge + Go paths.** Go's wasip1 path resolution only locates files under
  WasmEdge when the guest mount point is the filesystem root, so the WasmEdge
  golua target mounts at `/` and opens `/fib.lua` (other runtimes accept an
  arbitrary guest path like `/scripts`).

## Adding another runtime

For a Go-based runtime: create `runtimes/<name>/`, add a CLI under
`runtimes/<name>/cmd/<name>-run`, run `bazel run //:gazelle`, and provide a
`<name>_run` macro (see wazero).

For a prebuilt CLI runtime: add an `http_archive` to `MODULE.bazel` with a small
`<name>.BUILD` exposing the binary (and any runtime libs) via a `runtime_files`
filegroup, then add a thin `runtimes/<name>/defs.bzl` macro that calls
`cli_wasm_run` with the runtime's run/mount/arg conventions. Document it here and
ensure the runtime appears in the catalog.
