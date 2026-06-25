"""Bazel rule for compiling Go to wasm with the TinyGo toolchain.

TinyGo produces dramatically smaller wasm modules than the standard Go compiler
(tens of KB vs. ~1MB) by using its own LLVM-based backend and a slimmed-down
runtime/stdlib. The toolchain is fetched hermetically as a prebuilt release
(see the tinygo_* http_archive in MODULE.bazel); this rule drives it using
rules_go's Go SDK, so no system Go install is needed.

Usage (single package, stdlib only):

    load("//compilers/tinygo:defs.bzl", "tinygo_wasm")

    tinygo_wasm(
        name = "hello_tinygo_wasm",
        srcs = ["main.go"],
        target = "wasip1",   # wasip1 (WASI) or wasm (browser/JS host)
    )

With stdlib-only third-party dependencies, supply them via `deps`, each a
tinygo_go_package() naming its importpath and sources. The rule assembles a
GOPATH and builds in GO111MODULE=off mode (TinyGo's module support needs a
network-fetching go toolchain, which would break hermeticity).

The output is "<name>.wasm".
"""

load("@rules_go//go:def.bzl", "GoInfo")

GO_TOOLCHAIN = "@rules_go//go:toolchain"

TinyGoPackageInfo = provider(
    doc = "A Go package's importpath and source files for GOPATH assembly.",
    fields = {
        "importpath": "Go import path, e.g. github.com/Shopify/go-lua",
        "srcs": "depset of .go source File objects",
    },
)

def _tinygo_go_package_impl(ctx):
    # Pull importpath + sources straight from a rules_go go_library, so the
    # exact same (patched) sources used by the standard Go build feed TinyGo.
    gi = ctx.attr.lib[GoInfo]
    return [TinyGoPackageInfo(
        importpath = ctx.attr.importpath or gi.importpath,
        srcs = depset(gi.srcs),
    )]

tinygo_go_package = rule(
    implementation = _tinygo_go_package_impl,
    doc = "Adapts a rules_go go_library into a tinygo_wasm dep (importpath + srcs).",
    attrs = {
        "lib": attr.label(providers = [GoInfo], mandatory = True),
        "importpath": attr.string(
            doc = "Override import path; defaults to the library's importpath.",
        ),
    },
)

def _tinygo_wasm_impl(ctx):
    sdk = ctx.toolchains[GO_TOOLCHAIN].sdk
    tinygo_files = ctx.attr._tinygo_root.files.to_list()
    tinygo_bin = ctx.executable._tinygo

    out = ctx.actions.declare_file(ctx.label.name + ".wasm")

    # GOROOT is the directory containing the SDK root marker file.
    goroot = sdk.root_file.dirname

    # TinyGo distribution root: the bin/tinygo binary lives at <root>/bin/tinygo.
    tinygo_root = tinygo_bin.dirname[:-len("/bin")] if tinygo_bin.dirname.endswith("/bin") else tinygo_bin.dirname

    srcs = ctx.files.srcs
    src_args = [s.path for s in srcs]

    # Stage each dependency package's sources under GOPATH/src/<importpath>.
    stage_cmds = []
    dep_inputs = []
    for dep in ctx.attr.deps:
        info = dep[TinyGoPackageInfo]
        dst = "$GOPATH/src/" + info.importpath
        stage_cmds.append('mkdir -p "%s"' % dst)
        for f in info.srcs.to_list():
            dep_inputs.append(f)
            stage_cmds.append('cp "$(pwd)/%s" "%s/"' % (f.path, dst))
    staging = "\n".join(stage_cmds)

    # TinyGo shells out to the Go compiler and writes to a cache directory, so
    # we provide a writable HOME/cache inside the action tree and put the SDK's
    # go binary on PATH. With deps we build in GOPATH mode (GO111MODULE=off).
    module_mode = "off" if ctx.attr.deps else "auto"
    cache = out.path + ".cache"
    cmd = """
set -euo pipefail
export GOROOT="$(pwd)/{goroot}"
export PATH="$(pwd)/{goroot}/bin:$PATH"
export TINYGOROOT="$(pwd)/{tinygo_root}"
export HOME="$(pwd)/{cache}"
export XDG_CACHE_HOME="$(pwd)/{cache}/.cache"
export GOCACHE="$(pwd)/{cache}/.gocache"
export GOPATH="$(pwd)/{cache}/.gopath"
export GO111MODULE={module_mode}
mkdir -p "$HOME" "$XDG_CACHE_HOME" "$GOCACHE" "$GOPATH"
{staging}
exec "$(pwd)/{tinygo}" build -target={target} -o "{out}" {srcs}
""".format(
        goroot = goroot,
        tinygo_root = tinygo_root,
        tinygo = tinygo_bin.path,
        cache = cache,
        module_mode = module_mode,
        staging = staging,
        target = ctx.attr.target,
        out = out.path,
        srcs = " ".join(src_args),
    )

    ctx.actions.run_shell(
        outputs = [out],
        inputs = depset(
            direct = srcs + dep_inputs + [sdk.root_file, sdk.go],
            transitive = [sdk.libs, sdk.srcs, sdk.tools, sdk.headers, depset(tinygo_files)],
        ),
        command = cmd,
        mnemonic = "TinyGoBuild",
        progress_message = "Compiling %s to wasm with TinyGo" % ctx.label,
        use_default_shell_env = False,
        execution_requirements = {"no-sandbox": "1"},
    )

    return [DefaultInfo(files = depset([out]))]

tinygo_wasm = rule(
    implementation = _tinygo_wasm_impl,
    doc = "Compile Go sources to a wasm module using the TinyGo toolchain.",
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".go"],
            mandatory = True,
            doc = "Go source files to compile.",
        ),
        "target": attr.string(
            default = "wasip1",
            doc = "TinyGo target: 'wasip1' (WASI) or 'wasm' (JS host).",
        ),
        "deps": attr.label_list(
            providers = [TinyGoPackageInfo],
            doc = "Stdlib-only Go packages (tinygo_go_package) staged into GOPATH.",
        ),
        "_tinygo": attr.label(
            default = "@tinygo_linux_amd64//:tinygo_binary",
            executable = True,
            cfg = "exec",
            allow_single_file = True,
        ),
        "_tinygo_root": attr.label(
            default = "@tinygo_linux_amd64//:tinygo_root",
        ),
    },
    toolchains = [GO_TOOLCHAIN],
)
