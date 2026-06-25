# examples/

End-to-end demonstrations: source → wasm → run on a runtime.

- [hello-go-wasm](hello-go-wasm) — a Go program compiled to wasm (`wasip1`) by
  Bazel and executed on the wazero runtime. Run it:

      bazel run //examples/hello-go-wasm:run
