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
- [x] **Four more runtimes** as `bazel run` targets: Wasmtime, Wasmer,
      WasmEdge, and WAMR (`iwasm`), fetched as prebuilt CLIs and wrapped by a
      shared `cli_wasm_run` rule. Every hello/golua example now runs unmodified
      on all five runtimes, including host directory mounts.

- [x] **Third compiler: AssemblyScript** (`assemblyscript_wasm`). `asc` is
      JavaScript, so a prebuilt Node.js and the asc/binaryen/long/wasi-shim npm
      packages are fetched hermetically and assembled into a `node_modules/`
      tree per action. `examples/hello-assemblyscript-wasm` runs on all five
      runtimes.

## Next

- [ ] **More compilers → wasm**, each with a runnable example:
      Rust (`wasm32-wasip1`), Emscripten (C/C++).
- [ ] **Cross-platform runtime fetch**: the prebuilt runtime archives are
      currently linux/amd64 only; select per-OS/arch via `http_archive` +
      platform constraints (same gap as the TinyGo toolchain).
- [ ] **AOT-compile** modules where runtimes support it (`wasmedgec`,
      `wasmer compile`, `wamrc`) for faster startup as a separate Bazel action.
- [ ] **More interpreters in wasm**: a Lox or Scheme; richer Lua demos (read
      scripts from argv with FS mounts — already supported by `wazero-run -dir`).
- [ ] **Tools on wasm**: wabt/wasm-tools for inspect/validate/optimize; wire as
      Bazel actions.
- [ ] Generalize `tinygo_wasm` deps beyond stdlib-only (transitive
      `tinygo_go_package` generation).
- [ ] Catalog UX: a small report/query CLI and/or a static site generated from
      `data/catalog.json`.

## Toward self-hosting

The north star: run the compilers/interpreters in this repo *inside* wasm. The
catalog makes candidates easy to spot — a tool can run in wasm when its
**implementation language** (GitHub's `language` field) targets wasm. Tractable
today with our existing toolchains are the pure-Go entries (goja/otto = JS,
GopherLua = Lua, wa-lang = a Go compiler with a native wasm backend), plus
AssemblyScript.

- [ ] **`asc` in wasm** (the AssemblyScript self-hosting milestone). The
      frontend is portable JS, but the Binaryen backend calls the host's
      `WebAssembly.instantiate`, so a plain JS-engine-in-wasm (QuickJS, goja)
      can't host it. Needs a **WASI JS engine that implements the WebAssembly
      API** — realistically SpiderMonkey (StarlingMonkey / ComponentizeJS),
      which runs as a component (Wasmtime-only, breaking the 5-runtime
      symmetry). Spike the engine first.
- [x] **A Go compiler/interpreter running itself in wasm** — done two ways:
      `interpreters/wa` (the Wa toolchain — a Go compiler with a native wasm
      backend + embedded wazero — compiles *and* runs Wa from inside a runtime,
      and emits WAT via `build_hello_wat`), and `interpreters/goja` (a Go JS
      engine running JavaScript in wasm). Both reuse the golua pattern. The wa
      module is ~31 MB, so it's happy on wazero/Wasmtime/WasmEdge but slow on
      Wasmer / too big for WAMR's interpreter.
- [ ] Compile a wasm runtime itself to wasm (e.g. a Go/Rust runtime via its own
      wasm target) and run a guest module under the wasm-compiled runtime
      (runtime-in-runtime).
- [ ] WASI 0.2 / Component Model support where toolchains require it.

## Principles

- Hermetic and reproducible: a fresh machine with only Bazel can build, test,
  and run everything.
- The catalog is checked in; the repo is useful offline.
- Every capability ships with a runnable example and docs.
