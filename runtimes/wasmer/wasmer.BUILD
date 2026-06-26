# Build file for the prebuilt Wasmer release (wasmerio, Rust).
# Archive layout: bin/wasmer, lib/*.so (C API, unused by the CLI), include/.
package(default_visibility = ["//visibility:public"])

exports_files(["bin/wasmer"])

# The `wasmer` CLI links only against system libs (libstdc++, libm), so the
# binary alone suffices at runtime.
filegroup(
    name = "runtime_files",
    srcs = ["bin/wasmer"],
)
