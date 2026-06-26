"""Shared Bazel rule for running a .wasm module on an external CLI runtime.

The mature wasm runtimes in this repo (Wasmtime, Wasmer, WasmEdge, WAMR) ship as
prebuilt command-line tools rather than Go libraries. `cli_wasm_run` wraps any of
them in a hermetic, location-independent `bazel run` target, mirroring the
`wazero_run` macro's UX:

    bazel run //examples/hello-go-wasm:run_wasmtime -- extra args

The module, the runtime CLI, and the CLI's runtime files (e.g. WasmEdge's shared
library) are all carried as runfiles. The generated wrapper resolves them by
runfiles path so the target works from any directory.

Each runtime differs in how it spells "run this module" and "mount this
directory", so the per-runtime macros in //runtimes/<name>:defs.bzl supply:

  * run_prefix      args before the module (e.g. ["run"] for wasmtime/wasmer)
  * mount_flag      the mount option token (e.g. "--dir", "--map-dir=")
  * mount_value     a "{host}"/"{guest}" template for the mount argument
  * mount_inline    True to glue flag+value into one token (--map-dir=g::h)
  * args_separator  a token between module and its args (e.g. "--" for wasmer)

Mount HOST paths are resolved against runfiles at run time; place the files to
mount in `data` and reference them by their repo-relative path in `mounts`.
"""

def _cli_wasm_run_impl(ctx):
    module = ctx.file.module
    runner = ctx.file.runner

    # Per-runtime mount translation. Input is HOST[:GUEST].
    #
    # Bazel runfiles are symlinks pointing back into the source tree / bazel-out.
    # wazero follows them transparently, but the stricter WASI sandboxes
    # (Wasmtime, Wasmer, WasmEdge, WAMR) refuse to follow a symlink that escapes
    # a preopened directory. So we materialize each mounted directory into a
    # real, dereferenced copy under a temp dir at run time, and mount that.
    materialize = []
    mount_tokens = []
    for i, m in enumerate(ctx.attr.mounts):
        host, _, guest = m.partition(":")
        if not guest:
            guest = host
        mdir = "$_tmproot/%d" % i
        materialize.append('mkdir -p "{d}" && cp -RL "$base/{h}/." "{d}/"'.format(d = mdir, h = host))
        value = ctx.attr.mount_value.replace("{host}", mdir).replace("{guest}", guest)
        if ctx.attr.mount_inline:
            mount_tokens.append(ctx.attr.mount_flag + value)
        else:
            mount_tokens.append(ctx.attr.mount_flag)
            mount_tokens.append(value)

    sep = [ctx.attr.args_separator] if ctx.attr.args_separator else []

    pre_s = " ".join([repr(a) for a in ctx.attr.run_prefix])
    mounts_s = " ".join([repr(a) for a in mount_tokens])
    sep_s = " ".join([repr(a) for a in sep])
    extra_s = " ".join([repr(a) for a in ctx.attr.module_args])

    invoke = '"$base/$RUNNER" {pre} {mounts} "$base/$MODULE" {sep} {extra} "$@"'.format(
        pre = pre_s,
        mounts = mounts_s,
        sep = sep_s,
        extra = extra_s,
    )

    if materialize:
        # Set up a temp dir, populate the mounts, run (not exec, so the cleanup
        # trap fires), and propagate the runtime's exit code.
        run_block = """_tmproot="$(mktemp -d)"
trap 'rm -rf "$_tmproot"' EXIT
{materialize}
{invoke}
""".format(materialize = "\n".join(materialize), invoke = invoke)
    else:
        run_block = "exec " + invoke + "\n"

    script = ctx.actions.declare_file(ctx.label.name + ".sh")
    ctx.actions.write(
        output = script,
        is_executable = True,
        content = """#!/usr/bin/env bash
set -euo pipefail
# Resolve runfiles regardless of the invocation directory.
RUNFILES="${{RUNFILES_DIR:-${{0}}.runfiles}}"
RUNNER="{runner}"
MODULE="{module}"
# Under `bazel run`, the main repo's runfiles live in $RUNFILES/_main; external
# repos (the runtime CLIs) are siblings, reachable via the runner's ../ path.
if [[ -f "$RUNFILES/_main/$MODULE" ]]; then
  base="$RUNFILES/_main"
else
  base="$(dirname "$0")"
fi
{run_block}""".format(
            runner = runner.short_path,
            module = module.short_path,
            run_block = run_block,
        ),
    )

    runfiles = ctx.runfiles(
        files = [module, runner] + ctx.files.runner_files + ctx.files.data,
    )
    return [DefaultInfo(executable = script, runfiles = runfiles)]

cli_wasm_run = rule(
    implementation = _cli_wasm_run_impl,
    executable = True,
    doc = "Run a .wasm module on an external CLI wasm runtime as a `bazel run` target.",
    attrs = {
        "module": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "The wasm module to execute (a go_cross_binary output or a .wasm file).",
        ),
        "module_args": attr.string_list(
            doc = "Arguments passed to the module before any `bazel run -- ...` args.",
        ),
        "data": attr.label_list(
            allow_files = True,
            doc = "Extra files to include in runfiles (e.g. scripts to mount).",
        ),
        "mounts": attr.string_list(
            doc = "Directories to mount as HOST[:GUEST]; HOST is a runfiles path.",
        ),
        "runner": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "The runtime CLI binary (an external-repo file label).",
        ),
        "runner_files": attr.label_list(
            allow_files = True,
            doc = "Additional runtime files for the CLI (e.g. a shared library).",
        ),
        "run_prefix": attr.string_list(
            doc = "Args inserted before the module (subcommand and/or default flags).",
        ),
        "mount_flag": attr.string(
            doc = "The option token introducing a directory mount.",
        ),
        "mount_value": attr.string(
            doc = "A {host}/{guest} template producing the mount argument.",
        ),
        "mount_inline": attr.bool(
            default = False,
            doc = "If True, concatenate mount_flag and the value into a single token.",
        ),
        "args_separator": attr.string(
            doc = "A token placed between the module and its args (e.g. '--').",
        ),
    },
)
