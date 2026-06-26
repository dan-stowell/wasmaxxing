# Compilers (targeting wasm)

A "compiler" here is any toolchain that emits a `.wasm` module. We wire each up
as Bazel build targets so `bazel build` produces wasm artifacts hermetically.

## Go → wasm (rules_go, `wasip1`)

The Go toolchain (1.21+) compiles to `GOOS=wasip1 GOARCH=wasm`, producing a WASI
preview-1 module. rules_go exposes this via `go_cross_binary` plus a platform.

```python
load("@rules_go//go:def.bzl", "go_binary", "go_cross_binary", "go_library")

go_library(name = "app_lib", srcs = ["main.go"], importpath = "...")
go_binary(name = "app", embed = [":app_lib"])

go_cross_binary(
    name = "app_wasm",
    target = ":app",
    platform = "//platforms:wasip1_wasm",
)
```

```sh
bazel build //path/to:app_wasm
file bazel-bin/path/to/app_wasm   # WebAssembly (wasm) binary module
```

`//platforms:wasip1_wasm` carries the rules_go constraint values
(`@rules_go//go/toolchain:wasip1` + `:wasm`) that select the wasm toolchain.

Worked end-to-end example: [`examples/hello-go-wasm`](../examples/hello-go-wasm).

## Go → wasm (TinyGo)

[TinyGo](https://tinygo.org) is an alternative Go compiler with an LLVM backend
and a slimmed-down runtime/stdlib. It produces much smaller wasm modules than
standard Go (e.g. ~0.67 MB vs ~2.6 MB for hello-world; ~2.2 MB vs ~4.5 MB for
the Lua interpreter).

The toolchain is fetched hermetically as a prebuilt release (an `http_archive`
in `MODULE.bazel`) and driven by the `tinygo_wasm` rule in
[`compilers/tinygo/defs.bzl`](../compilers/tinygo/defs.bzl). It uses rules_go's
Go SDK — **no system Go install required**.

```python
load("//compilers/tinygo:defs.bzl", "tinygo_wasm")

tinygo_wasm(
    name = "hello_tinygo_wasm",
    srcs = ["main.go"],
    target = "wasip1",   # or "wasm" for a JS host
)
```

```sh
bazel build //examples/hello-tinygo-wasm:hello_tinygo_wasm
bazel run   //examples/hello-tinygo-wasm:run     # build + run on wazero
```

### Third-party dependencies

TinyGo's module mode would need a network-fetching `go` at build time, breaking
hermeticity. Instead, `tinygo_wasm` builds in GOPATH mode (`GO111MODULE=off`)
and stages dependency sources from rules_go `go_library` targets via
`tinygo_go_package`:

```python
load("//compilers/tinygo:defs.bzl", "tinygo_go_package", "tinygo_wasm")

tinygo_go_package(
    name = "go_lua_pkg",
    lib = "@com_github_shopify_go_lua//:go-lua",  # reuses the same sources
)

tinygo_wasm(
    name = "golua_tinygo_wasm",
    srcs = ["main.go"],
    deps = [":go_lua_pkg"],
)
```

This reuses the *exact* (patched) dependency sources from the standard build, so
both Go and TinyGo see identical code. Currently limited to stdlib-only
dependencies; transitive third-party deps would each need a `tinygo_go_package`.
Worked example: [`interpreters/golua`](../interpreters/golua) builds the Lua
interpreter with both standard Go and TinyGo.

## AssemblyScript → wasm (`asc`)

[AssemblyScript](https://www.assemblyscript.org) is a typed subset of TypeScript
that compiles to WebAssembly. Its compiler, `asc`, is distributed as JavaScript:
a portable frontend plus a [Binaryen](https://github.com/WebAssembly/binaryen)
backend that is *itself* a wasm module. So the toolchain has two hermetic parts,
both fetched as `http_archive`s in `MODULE.bazel` (**no system Node required**):

- a prebuilt **Node.js** runtime (`@nodejs_linux_amd64`) to execute `asc`, and
- the npm tarballs `assemblyscript`, `binaryen`, `long`, and
  `@assemblyscript/wasi-shim`.

The `assemblyscript_wasm` rule in
[`compilers/assemblyscript/defs.bzl`](../compilers/assemblyscript/defs.bzl)
assembles those packages into a `node_modules/` tree per build action (Node's
ESM resolver needs a real tree; `NODE_PATH` doesn't apply) and runs `asc`:

```python
load("//compilers/assemblyscript:defs.bzl", "assemblyscript_wasm")

assemblyscript_wasm(
    name = "hello_as_wasm",
    srcs = ["hello.ts"],
    entry = "hello.ts",
)
```

```sh
bazel build //examples/hello-assemblyscript-wasm:hello_as_wasm
bazel run   //examples/hello-assemblyscript-wasm:run            # on wazero
bazel run   //examples/hello-assemblyscript-wasm:run_wasmtime   # ...or any of 5
```

By default the module is built as a WASI **command** via
`@assemblyscript/wasi-shim`, so it has a `_start` and can `console.log` — it runs
unmodified on all five runtimes, like the Go/TinyGo examples. Worked example:
[`examples/hello-assemblyscript-wasm`](../examples/hello-assemblyscript-wasm).

> **Toward self-hosting.** `asc` is the canonical "compiler that targets wasm",
> but running `asc` *itself* in wasm is gated by its Binaryen backend, which
> calls the host's `WebAssembly.instantiate`. A simple JS engine in wasm
> (QuickJS, goja) can't host it; that needs a WASI JS engine implementing the
> WebAssembly API (e.g. SpiderMonkey). See the [roadmap](roadmap.md).

## Adding another compiler

1. Pick a target from the catalog (`kind == "compiler"`). High-value,
   Bazel-friendly options include **TinyGo** (smaller modules), **Rust**
   (`wasm32-wasip1`), **AssemblyScript**, and **Emscripten** (C/C++).
2. Add Bazel rules under `compilers/<name>/` (or reuse existing toolchain rules
   like rules_rust). Aim for a macro that turns sources into a `.wasm`.
3. Add an `examples/` target that builds *and* runs the result on a runtime
   (`wazero_run`), giving a full source → wasm → run loop.
4. Document the compiler here.

## Notes on WASI versions

The Go `wasip1` target and wazero's `wasi_snapshot_preview1` import speak the
same ABI, so Go modules run unmodified. Other toolchains may target
`wasm32-unknown-unknown` (no WASI) or the newer Component Model / WASI 0.2 —
those need different host wiring, tracked in the [roadmap](roadmap.md).
