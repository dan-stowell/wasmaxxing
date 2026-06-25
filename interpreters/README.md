# interpreters/

Interpreters compiled to wasm, so scripts can be run *inside* a wasm runtime.

- [golua](golua) — a Lua 5.2 interpreter (Shopify/go-lua) compiled to wasm with
  both standard Go and TinyGo, run on wazero.

See [../docs/interpreters.md](../docs/interpreters.md).
