# Generic BUILD for an npm package fetched as an http_archive (strip_prefix
# "package"). Exposes the whole package tree as a filegroup plus its package.json
# (a stable sentinel the assemblyscript_wasm rule uses to locate the package
# root when assembling a node_modules/ layout).
package(default_visibility = ["//visibility:public"])

exports_files(["package.json"])

filegroup(
    name = "files",
    srcs = glob(["**"], exclude = ["BUILD.bazel", "WORKSPACE", "WORKSPACE.bazel"]),
)
