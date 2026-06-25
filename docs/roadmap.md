# Roadmap

The north star: **self-hosting** — compile the compilers, tools, and runtimes in
this repo to wasm and run them on each other. We get there incrementally.

## Done

- [x] Bazel monorepo bootstrap (bzlmod, rules_go, gazelle), hermetic builds.
- [x] Data pipeline: parse the four seed lists → normalized `data/catalog.json`
      (500+ entries); optional GitHub enrichment.
- [x] First runtime: **wazero** (pure Go) with a `bazel run` CLI and macro.
- [x] First compiler path: **Go → wasm** (`wasip1`) via `go_cross_binary`.
- [x] First end-to-end slice: `examples/hello-go-wasm` compiled to wasm and run
      on wazero, all through `bazel run`.
- [x] Second compiler: **TinyGo** (`tinygo_wasm` rule), fetched hermetically,
      producing much smaller modules. Example: `examples/hello-tinygo-wasm`.
- [x] **Run an interpreter in wasm**: `interpreters/golua`, a Lua 5.2 VM
      (go-lua) compiled to wasm with both standard Go and TinyGo, running
      scripts on wazero.

## Next

- [ ] **More compilers → wasm**, each with a runnable example:
      Rust (`wasm32-wasip1`), AssemblyScript, Emscripten (C/C++).
- [ ] **More runtimes** as `bazel run` targets: Wasmtime, Wasmer, WasmEdge,
      WAMR. Track which need system toolchains vs. fetch prebuilt.
- [ ] **More interpreters in wasm**: a Lox or Scheme; richer Lua demos (read
      scripts from argv with FS mounts — already supported by `wazero-run -dir`).
- [ ] **Tools on wasm**: wabt/wasm-tools for inspect/validate/optimize; wire as
      Bazel actions.
- [ ] Generalize `tinygo_wasm` deps beyond stdlib-only (transitive
      `tinygo_go_package` generation).
- [ ] Catalog UX: a small report/query CLI and/or a static site generated from
      `data/catalog.json`.

## Toward self-hosting

- [ ] Compile a wasm runtime itself to wasm (e.g. a Go/Rust runtime via its own
      wasm target) and run a guest module under the wasm-compiled runtime
      (runtime-in-runtime).
- [ ] Compile a compiler to wasm and use it to produce wasm from inside a
      runtime.
- [ ] WASI 0.2 / Component Model support where toolchains require it.

## Principles

- Hermetic and reproducible: a fresh machine with only Bazel can build, test,
  and run everything.
- The catalog is checked in; the repo is useful offline.
- Every capability ships with a runnable example and docs.
