# examples/

End-to-end demonstrations: source → wasm → run on a runtime.

- [hello-go-wasm](hello-go-wasm) — a Go program compiled to wasm (`wasip1`) by
  the standard Go compiler and executed on the wazero runtime:

      bazel run //examples/hello-go-wasm:run

- [hello-tinygo-wasm](hello-tinygo-wasm) — the same idea, compiled with TinyGo
  for a much smaller module:

      bazel run //examples/hello-tinygo-wasm:run

See also [interpreters/golua](../interpreters/golua) for a Lua interpreter
compiled to wasm.
