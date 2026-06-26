# Build file for the prebuilt WAMR `iwasm` release
# (WebAssembly Micro Runtime, Bytecode Alliance, C).
# Archive layout: a single `iwasm` binary at the root.
package(default_visibility = ["//visibility:public"])

exports_files(["iwasm"])

# iwasm links only against system libs (libz, libzstd); the binary alone is
# enough at runtime.
filegroup(
    name = "runtime_files",
    srcs = ["iwasm"],
)
