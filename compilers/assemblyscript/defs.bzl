"""Bazel rule for compiling AssemblyScript to wasm with `asc`.

AssemblyScript is a typed subset of TypeScript that compiles to WebAssembly. Its
compiler, `asc`, is distributed as JavaScript (a portable frontend plus a
Binaryen backend that is itself wasm), so we drive it with a hermetically
fetched Node.js runtime (no system Node required). The compiler bundle and its
runtime deps are fetched as npm tarballs (see the npm_* http_archives in
MODULE.bazel) and assembled into a node_modules/ tree per build action.

Usage:

    load("//compilers/assemblyscript:defs.bzl", "assemblyscript_wasm")

    assemblyscript_wasm(
        name = "hello_as_wasm",
        srcs = ["hello.ts"],
        entry = "hello.ts",
    )

By default the module is built as a WASI *command* using @assemblyscript/wasi-
shim, so it has a `_start` and can `console.log` -- i.e. it runs unmodified on
wazero, Wasmtime, Wasmer, WasmEdge and WAMR, just like the Go/TinyGo examples.

The output is "<name>.wasm".
"""

def _assemblyscript_wasm_impl(ctx):
    node = ctx.executable._node
    out = ctx.actions.declare_file(ctx.label.name + ".wasm")

    # Each npm package's root is the directory holding its package.json. We copy
    # those trees into a node_modules/ layout so asc resolves bare imports
    # (binaryen, long) and the WASI shim's asconfig the way Node expects.
    asc_root = ctx.file._asc_pkg.dirname
    binaryen_root = ctx.file._binaryen_pkg.dirname
    long_root = ctx.file._long_pkg.dirname
    shim_root = ctx.file._shim_pkg.dirname

    entry = ctx.file.entry
    if entry == None:
        if len(ctx.files.srcs) != 1:
            fail("assemblyscript_wasm: set `entry` when there is more than one src")
        entry = ctx.files.srcs[0]

    optimize = ["--optimize"] if ctx.attr.optimize else []

    # Stage a node_modules/ tree, then run asc from inside it (so Node resolves
    # the bare `binaryen`/`long` imports and the shim's asconfig as it expects),
    # passing the entry and output as absolute execroot paths.
    cmd = """
set -euo pipefail
ROOT="$(pwd)"
W="$ROOT/{out}.work"
rm -rf "$W"
mkdir -p "$W/node_modules/@assemblyscript"
cp -RL "{asc_root}/." "$W/node_modules/assemblyscript/"
cp -RL "{binaryen_root}/." "$W/node_modules/binaryen/"
cp -RL "{long_root}/." "$W/node_modules/long/"
cp -RL "{shim_root}/." "$W/node_modules/@assemblyscript/wasi-shim/"
export HOME="$W/.home"
mkdir -p "$HOME"
cd "$W"
"$ROOT/{node}" node_modules/assemblyscript/bin/asc.js \\
  "$ROOT/{entry}" {opt} \\
  --config node_modules/@assemblyscript/wasi-shim/asconfig.json \\
  --outFile "$ROOT/{out}"
""".format(
        out = out.path,
        node = node.path,
        asc_root = asc_root,
        binaryen_root = binaryen_root,
        long_root = long_root,
        shim_root = shim_root,
        entry = entry.path,
        opt = " ".join(optimize),
    )

    inputs = depset(
        direct = ctx.files.srcs,
        transitive = [
            ctx.attr._asc_files.files,
            ctx.attr._binaryen_files.files,
            ctx.attr._long_files.files,
            ctx.attr._shim_files.files,
            ctx.attr._node.files,
        ],
    )

    ctx.actions.run_shell(
        outputs = [out],
        inputs = inputs,
        command = cmd,
        mnemonic = "AscBuild",
        progress_message = "Compiling %s to wasm with AssemblyScript" % ctx.label,
        use_default_shell_env = False,
        execution_requirements = {"no-sandbox": "1"},
    )

    return [DefaultInfo(files = depset([out]))]

assemblyscript_wasm = rule(
    implementation = _assemblyscript_wasm_impl,
    doc = "Compile AssemblyScript sources to a WASI wasm module using asc.",
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".ts"],
            mandatory = True,
            doc = "AssemblyScript (.ts) source files.",
        ),
        "entry": attr.label(
            allow_single_file = [".ts"],
            doc = "The entry module; required when there is more than one src.",
        ),
        "optimize": attr.bool(
            default = True,
            doc = "Pass --optimize to asc.",
        ),
        "_node": attr.label(
            default = "@nodejs_linux_amd64//:bin/node",
            executable = True,
            cfg = "exec",
            allow_single_file = True,
        ),
        "_asc_files": attr.label(default = "@npm_assemblyscript//:files"),
        "_asc_pkg": attr.label(default = "@npm_assemblyscript//:package.json", allow_single_file = True),
        "_binaryen_files": attr.label(default = "@npm_binaryen//:files"),
        "_binaryen_pkg": attr.label(default = "@npm_binaryen//:package.json", allow_single_file = True),
        "_long_files": attr.label(default = "@npm_long//:files"),
        "_long_pkg": attr.label(default = "@npm_long//:package.json", allow_single_file = True),
        "_shim_files": attr.label(default = "@npm_assemblyscript_wasi_shim//:files"),
        "_shim_pkg": attr.label(default = "@npm_assemblyscript_wasi_shim//:package.json", allow_single_file = True),
    },
)
