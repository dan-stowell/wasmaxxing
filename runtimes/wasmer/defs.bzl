"""`wasmer_run`: run a .wasm module on Wasmer as a `bazel run` target.

    load("//runtimes/wasmer:defs.bzl", "wasmer_run")

    wasmer_run(
        name = "run_wasmer",
        module = ":hello_wasm",
        module_args = ["world"],
    )

Wasmer spells `wasmer run [--volume HOST:GUEST] MODULE -- [args...]`; module
arguments go after a `--` separator.
"""

load("//runtimes:cli_run.bzl", "cli_wasm_run")

def wasmer_run(name, module, module_args = [], data = [], mounts = [], **kwargs):
    cli_wasm_run(
        name = name,
        module = module,
        module_args = module_args,
        data = data,
        mounts = mounts,
        runner = "@wasmer//:bin/wasmer",
        runner_files = ["@wasmer//:runtime_files"],
        run_prefix = ["run"],
        mount_flag = "--volume",
        mount_value = "{host}:{guest}",
        args_separator = "--",
        **kwargs
    )
