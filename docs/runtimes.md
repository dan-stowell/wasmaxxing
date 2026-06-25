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

## Adding another runtime

1. Create `runtimes/<name>/`.
2. If it's Go-based, add a CLI under `runtimes/<name>/cmd/<name>-run` and run
   `bazel run //:gazelle`. For other languages, wire up the appropriate Bazel
   rules.
3. Optionally provide a `<name>_run` macro mirroring wazero's for a consistent
   `bazel run` UX.
4. Document it here and ensure the runtime appears in the catalog.

Candidates from the catalog (by popularity): Wasmtime, Wasmer, WasmEdge, WAMR.
These pull in C/C++/Rust toolchains, so they come after the pure-Go baseline.
