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

# Compile a Go program to wasm and run it on the wazero runtime.
bazel run //examples/hello-go-wasm:run

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
| [`runtimes/`](runtimes/) | wasm runtimes wrapped as `bazel run` targets (currently **wazero**). |
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
- [docs/compilers.md](docs/compilers.md) — compiling to wasm.
- [docs/roadmap.md](docs/roadmap.md) — where this is going.

## Status

First vertical slice is live: a Go program is **compiled to wasm by Bazel**
(`go_cross_binary`, `GOOS=wasip1`) and **executed on the wazero runtime**, all
via `bazel run`. The data pipeline catalogs **500+** ecosystem entries.
