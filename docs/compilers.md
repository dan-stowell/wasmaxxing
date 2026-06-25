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
