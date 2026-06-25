# compilers/

Toolchains that emit wasm, wired up as Bazel rules/macros.

- Go → wasm via rules_go `go_cross_binary` (see [../platforms](../platforms)).
- [tinygo](tinygo) — the TinyGo toolchain (`tinygo_wasm` rule), fetched
  hermetically; produces much smaller modules.

See [../docs/compilers.md](../docs/compilers.md).
