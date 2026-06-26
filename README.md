# wasmaxxing

A Bazel monorepo for the WebAssembly ecosystem: **compilers that target wasm,
interpreters that run in wasm, tools that operate on wasm, and wasm runtimes** —
with a data pipeline that enumerates the whole landscape.

The long-term goal is *self-hosting*: building the compilers, tools, and
runtimes themselves as wasm. We're starting smaller — see
[the roadmap](docs/roadmap.md).

> This repo is the source of truth. Everything is documented here and built,
> tested, and run through Bazel. Clone onto a fresh machine with a working
> Bazel and you can build/test/run everything.

## Quick start

```sh
# Build & test everything.
bazel test //...

# Compile a Go program to wasm (standard Go) and run it on the wazero runtime.
bazel run //examples/hello-go-wasm:run

# Run the very same module on four other mature runtimes — fetched hermetically.
bazel run //examples/hello-go-wasm:run_wasmtime   # Wasmtime  (Rust)
bazel run //examples/hello-go-wasm:run_wasmer     # Wasmer    (Rust)
bazel run //examples/hello-go-wasm:run_wasmedge   # WasmEdge  (C++)
bazel run //examples/hello-go-wasm:run_wamr       # WAMR iwasm (C)

# Same, but compiled with TinyGo (much smaller module).
bazel run //examples/hello-tinygo-wasm:run

# Compile AssemblyScript (a typed TS subset) to wasm with asc (hermetic Node).
bazel run //examples/hello-assemblyscript-wasm:run

# Run a Lua interpreter that is itself compiled to wasm.
bazel run //interpreters/golua:run_fib
bazel run //interpreters/golua:run_fib_tinygo     # TinyGo build
bazel run //interpreters/golua:run_fib_wasmtime   # ...also on Wasmtime, etc.

# Regenerate the ecosystem catalog from the seed lists.
bazel run //pipeline/cmd/build-catalog
```

The only prerequisite is **Bazel** (via [bazelisk](https://github.com/bazelbuild/bazelisk);
the pinned version is in [`.bazelversion`](.bazelversion)). The Go SDK, the
wazero runtime, and all other dependencies are fetched hermetically by Bazel.
A C/C++ toolchain is only needed for the (future) C/C++-based pieces.

## Repository layout

| Path | Contents |
|------|----------|
| [`pipeline/`](pipeline/) | Data pipeline: parses the seed "awesome" lists and enriches them with GitHub metadata into [`data/catalog.json`](data/catalog.json). |
| [`data/`](data/) | Committed seed sources and the generated catalog. |
| [`runtimes/`](runtimes/) | wasm runtimes wrapped as `bazel run` targets: **wazero**, **Wasmtime**, **Wasmer**, **WasmEdge**, **WAMR**. |
| [`compilers/`](compilers/) | Toolchains that emit wasm (Bazel rules/macros). |
| [`interpreters/`](interpreters/) | Interpreters that run *in* wasm. |
| [`tools/`](tools/) | Tools that operate on wasm modules. |
| [`examples/`](examples/) | End-to-end examples: source → wasm → run. |
| [`platforms/`](platforms/) | Bazel platform definitions (e.g. `wasip1_wasm`). |
| [`docs/`](docs/) | Architecture, roadmap, and how-to docs. |

## Documentation

- [docs/architecture.md](docs/architecture.md) — how the pieces fit together.
- [docs/pipeline.md](docs/pipeline.md) — the data pipeline & catalog schema.
- [docs/runtimes.md](docs/runtimes.md) — running wasm modules.
- [docs/compilers.md](docs/compilers.md) — compiling to wasm (Go, TinyGo).
- [docs/interpreters.md](docs/interpreters.md) — interpreters running in wasm (Lua).
- [docs/roadmap.md](docs/roadmap.md) — where this is going.

## Status

Live vertical slices, all via `bazel run`:

- A Go program **compiled to wasm by Bazel** (`go_cross_binary`, `GOOS=wasip1`)
  and run on **five runtimes** — wazero, Wasmtime, Wasmer, WasmEdge, and WAMR —
  the latter four fetched hermetically as prebuilt CLIs.
- The same with **TinyGo** (`tinygo_wasm`, fetched hermetically) — ~4× smaller.
- **AssemblyScript → wasm** via `asc`, driven by a hermetically fetched Node.js,
  emitting a WASI command that also runs on all five runtimes.
- A **Lua interpreter compiled to wasm** (go-lua) running scripts on every
  runtime (with the host script directory mounted), built with both standard Go
  and TinyGo.

The data pipeline catalogs **500+** ecosystem entries.
