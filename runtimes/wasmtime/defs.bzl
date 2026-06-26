"""`wasmtime_run`: run a .wasm module on Wasmtime as a `bazel run` target.

    load("//runtimes/wasmtime:defs.bzl", "wasmtime_run")

    wasmtime_run(
        name = "run_wasmtime",
        module = ":hello_wasm",
        module_args = ["world"],
    )

Wasmtime spells `wasmtime run [--dir HOST::GUEST] MODULE [args...]`.
"""

load("//runtimes:cli_run.bzl", "cli_wasm_run")

def wasmtime_run(name, module, module_args = [], data = [], mounts = [], **kwargs):
    cli_wasm_run(
        name = name,
        module = module,
        module_args = module_args,
        data = data,
        mounts = mounts,
        runner = "@wasmtime//:wasmtime",
        runner_files = ["@wasmtime//:runtime_files"],
        run_prefix = ["run"],
        mount_flag = "--dir",
        mount_value = "{host}::{guest}",
        **kwargs
    )
