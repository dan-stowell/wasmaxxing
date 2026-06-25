# Architecture

wasmaxxing is a single Bazel module (bzlmod) organized around four concerns from
the WebAssembly ecosystem, plus a data pipeline that maps the territory.

```
            seed "awesome" lists (markdown, committed under data/sources/)
                                  |
                          pipeline/parse  (per-format parsers)
                                  |
                       pipeline/catalog   (normalize, dedupe, merge)
                                  |
                     data/catalog.json  <----  pipeline/github (enrich w/ GitHub API)
                                  |
               (drives which runtimes/compilers/projects we wire up)
                                  |
   +------------------+   +----------------+   +----------------+
   |  compilers/      |   |  runtimes/     |   |  interpreters/ |
   |  emit wasm       |   |  execute wasm  |   |  run in wasm   |
   +------------------+   +----------------+   +----------------+
            \                    |                    /
             \------------  examples/  --------------/
               source -> .wasm -> run on a runtime
```

## Build system

- **Bazel** with bzlmod (`MODULE.bazel`). Version pinned in `.bazelversion`.
- **rules_go + gazelle** provide the Go toolchain, cross-compilation to
  `wasip1/wasm`, and auto-generated `BUILD.bazel` files.
- Everything is hermetic: the Go SDK and third-party modules (e.g. wazero) are
  downloaded by Bazel, not the system. Cloning onto a fresh box with only
  Bazel installed is sufficient to `bazel build/test/run //...`.

## Components

### Data pipeline (`pipeline/`)
Pure-Go, fully unit-tested offline. Three packages:
- `catalog` — the normalized data model (`Entry`, `Kind`, dedupe/merge, JSON I/O).
- `parse` — format-specific parsers for the three seed list layouts.
- `github` — optional enrichment via the GitHub REST API.

Two CLIs under `pipeline/cmd/` regenerate and enrich `data/catalog.json`.
See [pipeline.md](pipeline.md).

### Runtimes (`runtimes/`)
Each runtime is wrapped so a wasm module can be executed with `bazel run`. The
first is **wazero** (pure Go, zero system dependencies), exposing:
- `runtimes/wazero/cmd/wazero-run` — a WASI preview-1 CLI runner.
- `runtimes/wazero/defs.bzl` — the `wazero_run()` macro.

See [runtimes.md](runtimes.md).

### Compilers (`compilers/`) and examples (`examples/`)
Compilers that target wasm are wired up as Bazel rules/macros:
- standard Go via rules_go `go_cross_binary` (`//platforms:wasip1_wasm`).
- **TinyGo** via the `tinygo_wasm` rule in `compilers/tinygo`, with the
  toolchain fetched hermetically as a prebuilt release and driven by rules_go's
  Go SDK (no system Go).

The `examples/` tree demonstrates full source → wasm → run pipelines
(`hello-go-wasm`, `hello-tinygo-wasm`). See [compilers.md](compilers.md).

### Interpreters (`interpreters/`)
Interpreters compiled to wasm, run on a runtime. `interpreters/golua` is a Lua
5.2 VM (go-lua) built with both standard Go and TinyGo. See
[interpreters.md](interpreters.md).

### Platforms (`platforms/`)
Reusable Bazel `platform()` targets. `//platforms:wasip1_wasm` selects the Go
`GOOS=wasip1 GOARCH=wasm` toolchain.

## Conventions

- Go import path prefix: `github.com/dan-stowell/wasmaxxing`.
- Run `bazel run //:gazelle` after adding/moving Go files to regenerate BUILD
  files. Hand-written non-Go targets (macros, platforms) coexist with gazelle
  output in the same BUILD files.
- Commit early and often; the catalog JSON is checked in so the repo is useful
  without network access.
