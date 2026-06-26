# Build file for the prebuilt Wasmtime release (Bytecode Alliance, Rust).
# The tarball strips to a single directory containing the `wasmtime` CLI.
package(default_visibility = ["//visibility:public"])

exports_files(["wasmtime"])

# Everything the CLI needs at runtime (here, just the binary itself; Wasmtime
# links only against system libs). Carried as runfiles by cli_wasm_run.
filegroup(
    name = "runtime_files",
    srcs = ["wasmtime"],
)
