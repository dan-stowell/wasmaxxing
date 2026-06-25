# compilers/

Toolchains that emit wasm, wired up as Bazel rules/macros. See
[../docs/compilers.md](../docs/compilers.md).

The first compiler path (Go → wasm via rules_go) lives in
[../examples/hello-go-wasm](../examples/hello-go-wasm) and
[../platforms](../platforms). Language-specific toolchains added here over time.
