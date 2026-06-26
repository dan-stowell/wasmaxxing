# Interpreters (running in wasm)

Interpreters and compilers compiled to wasm let you run scripts — and even
compile programs — *inside* a wasm runtime. With [`wa`](#wa--a-wasm-targeting-compiler-running-in-wasm-self-hosting)
this reaches genuine self-hosting: a wasm-targeting compiler running in wasm.

## golua — a Lua interpreter on wasm

[`interpreters/golua`](../interpreters/golua) is a small Lua 5.2 interpreter
built on [Shopify/go-lua](https://github.com/Shopify/go-lua), a pure-Go port of
the reference VM. Being pure Go (no cgo), it cross-compiles to wasm unchanged.

### Why go-lua (not gopher-lua)

The more popular [gopher-lua](https://github.com/yuin/gopher-lua) miscompiles
multiple-assignment: `x, y = y, x` yields `9 9` instead of `9 5`, because the
right-hand side isn't fully evaluated before assignment. go-lua follows the
reference semantics correctly. A regression test
([`main_test.go`](../interpreters/golua/cmd/golua/main_test.go)) guards this.

### Building & running

```sh
# Standard Go -> wasm, run fib.lua on wazero:
bazel run //interpreters/golua:run_fib

# TinyGo -> wasm (about half the size), same demo:
bazel run //interpreters/golua:run_fib_tinygo

# The same interpreter + mounted script on the other runtimes:
bazel run //interpreters/golua:run_fib_wasmtime   # also _wasmer, _wasmedge, _wamr

# Run your own snippet through the wasm interpreter via stdin:
echo 'print("hi", 6*7)' | \
  $(bazel cquery --output=files //runtimes/wazero/cmd/wazero-run) \
  $(bazel cquery --output=files //interpreters/golua/cmd/golua:golua_wasm)
```

The interpreter accepts a script file argument, `-e CODE`, or a script on
stdin, and exposes CLI args to Lua as the global `arg` table.

### wasm portability note

go-lua's `os.clock()` uses `syscall.Getrusage`, which doesn't exist on the
`wasip1`/`js` targets. A small hermetic patch
([`patches/go-lua-wasm.patch`](../patches/go-lua-wasm.patch), applied via
`go_deps.module_override` in `MODULE.bazel`) swaps in a wall-clock fallback on
wasm. See [pipeline.md](pipeline.md) and [compilers.md](compilers.md).

## goja — a JavaScript engine on wasm

[`interpreters/goja`](../interpreters/goja) wraps
[dop251/goja](https://github.com/dop251/goja), a pure-Go ECMAScript 5.1+ engine.
Compiled to wasm, it runs JavaScript from *inside* a WebAssembly module. The CLI
provides a minimal host (`console.log`/`print`, an `argv` global) and runs a
file, a `-e` snippet, or stdin:

```sh
bazel run //interpreters/goja:run_demo            # demo.js on wazero
bazel run //interpreters/goja:run_demo_wasmtime   # ...or any of the five runtimes
```

The bundled `demo.js` runs unmodified on all five runtimes.

## wa — a wasm-targeting compiler running *in* wasm (self-hosting)

[`interpreters/wa`](../interpreters/wa) is the milestone: the
[Wa language](https://wa-lang.org) toolchain (`wa-lang.org/wa`) is a pure-Go
compiler with a **native wasm backend** plus a small embedded wazero runtime, so
the whole thing cross-compiles to `wasip1`. A thin driver over its public `api`
package compiles **and** runs Wa programs — meaning a compiler that targets
WebAssembly is itself executing inside a WebAssembly runtime (wasm compiling and
running wasm, no nested JS engine required, unlike `asc`).

```sh
bazel run //interpreters/wa:run_hello         # compile + run hello.wa, in wasm
bazel run //interpreters/wa:build_hello_wat   # emit the compiled WAT, in wasm
```

The toolchain module is ~15 MB (a full compiler + runtime), which runs
comfortably on **wazero, Wasmtime and WasmEdge**. Wasmer's ahead-of-time
compiler is very slow on a module this size, and WAMR's loader can't handle its
~2800 functions, so those targets exist for parity but aren't the happy path.

A TinyGo build was investigated to shrink it, but doesn't work: TinyGo compiles
the toolchain (~11 MB) yet the result crashes at runtime — its conservative GC
faults in `runtime.scanConservative` during the compiler's heavy string
building, and `-gc=leaking` miscompiles wa's WAT parser. So wa stays on the
standard Go compiler, which produces a correct module.

## Module sizes

| Build | hello-world | Lua interpreter |
|-------|-------------|-----------------|
| standard Go (`go_cross_binary`) | ~2.6 MB | ~4.5 MB |
| TinyGo (`tinygo_wasm`) | ~0.67 MB | ~2.2 MB |

## Adding another interpreter

1. Find a pure-Go (or otherwise wasm-friendly) interpreter in the catalog.
2. Add a CLI under `interpreters/<name>/cmd/<name>` and a `go_cross_binary`
   wasm target; optionally a `tinygo_wasm` target.
3. Add a `wazero_run` demo and a unit test, then document it here.
