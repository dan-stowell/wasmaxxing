# Interpreters (running in wasm)

Interpreters compiled to wasm let you run scripts *inside* a wasm runtime — a
step toward self-hosting language tools.

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
