# Build file for the prebuilt WasmEdge release (CNCF, C++).
# Archive layout: bin/wasmedge, lib64/libwasmedge.so* (the CLI's RUNPATH is
# $ORIGIN/../lib64, so the binary and lib64/ must stay in their relative
# positions -- which http_archive + runfiles preserve).
package(default_visibility = ["//visibility:public"])

exports_files(["bin/wasmedge"])

# Carry the CLI together with the shared library it loads via RUNPATH.
filegroup(
    name = "runtime_files",
    srcs = [
        "bin/wasmedge",
    ] + glob(["lib64/libwasmedge.so*"]),
)
