"""Bazel helpers for running wasm modules on the wazero runtime.

`wazero_run` wraps a .wasm module in an executable `bazel run` target. The
module and the wazero-run binary are carried as runfiles, so the target is
hermetic and location-independent:

    load("//runtimes/wazero:defs.bzl", "wazero_run")

    wazero_run(
        name = "run_hello",
        module = "//examples/hello-go-wasm:hello_wasm",
        module_args = ["world"],
    )

Then: `bazel run //examples/hello-go-wasm:run_hello`.

To run a guest that reads files, mount a directory and add the files to data:

    wazero_run(
        name = "run_script",
        module = "//interpreters/golua/cmd/golua:golua_wasm",
        data = ["examples/fib.lua"],
        mounts = ["interpreters/golua/examples:/scripts"],
        module_args = ["/scripts/fib.lua"],
    )
"""

def _wazero_run_impl(ctx):
    module = ctx.file.module
    runner = ctx.executable._runner

    # -dir HOST:GUEST mounts, with HOST resolved against runfiles at runtime.
    mount_args = []
    for m in ctx.attr.mounts:
        host, _, guest = m.partition(":")
        if not guest:
            guest = host
        mount_args.append("-dir")
        mount_args.append("$base/" + host + ":" + guest)
    mounts = " ".join([repr(a) for a in mount_args])

    # Build a wrapper script that invokes the runner on the module via its
    # runfiles path, forwarding extra command-line arguments.
    script = ctx.actions.declare_file(ctx.label.name + ".sh")
    extra = " ".join([repr(a) for a in ctx.attr.module_args])
    ctx.actions.write(
        output = script,
        is_executable = True,
        content = """#!/usr/bin/env bash
set -euo pipefail
# Resolve runfiles regardless of invocation directory.
RUNFILES="${{RUNFILES_DIR:-${{0}}.runfiles}}"
RUNNER="{runner}"
MODULE="{module}"
# When run via `bazel run`, runfiles are siblings of $0.
if [[ -f "$RUNFILES/_main/$RUNNER" ]]; then
  base="$RUNFILES/_main"
else
  base="$(dirname "$0")"
fi
exec "$base/$RUNNER" {mounts} "$base/$MODULE" {extra} "$@"
""".format(
            runner = runner.short_path,
            module = module.short_path,
            mounts = mounts,
            extra = extra,
        ),
    )

    runfiles = ctx.runfiles(files = [module] + ctx.files.data)
    runfiles = runfiles.merge(ctx.attr._runner[DefaultInfo].default_runfiles)
    return [DefaultInfo(executable = script, runfiles = runfiles)]

wazero_run = rule(
    implementation = _wazero_run_impl,
    executable = True,
    doc = "Run a .wasm module on the wazero runtime as a `bazel run` target.",
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
        "_runner": attr.label(
            default = "//runtimes/wazero/cmd/wazero-run",
            executable = True,
            cfg = "target",
        ),
    },
)
