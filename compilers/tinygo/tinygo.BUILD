# BUILD overlay for the prebuilt TinyGo release archive.
# Exposes the whole distribution tree plus the tinygo driver binary.

package(default_visibility = ["//visibility:public"])

filegroup(
    name = "tinygo_binary",
    srcs = ["bin/tinygo"],
)

# Everything TinyGo needs at runtime: the driver, bundled clang/LLVM/wasm-opt,
# target definitions, and the TinyGo standard library sources.
filegroup(
    name = "tinygo_root",
    srcs = glob(
        ["**"],
        exclude = ["**/* *"],  # skip any paths with spaces (none expected)
    ),
)
