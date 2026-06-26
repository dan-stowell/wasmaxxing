"""`wasmedge_run`: run a .wasm module on WasmEdge as a `bazel run` target.

    load("//runtimes/wasmedge:defs.bzl", "wasmedge_run")

    wasmedge_run(
        name = "run_wasmedge",
        module = ":hello_wasm",
        module_args = ["world"],
    )

WasmEdge spells `wasmedge [--dir GUEST:HOST] MODULE [args...]` -- note the mount
order is guest-first, the opposite of Wasmtime.

Caveat: Go's wasip1 path resolution only finds files under WasmEdge when the
guest mount point is the filesystem root `/` (e.g. mounts = ["scripts:/"]);
other guest paths fail to open. Other runtimes accept arbitrary guest paths.
"""

load("//runtimes:cli_run.bzl", "cli_wasm_run")

def wasmedge_run(name, module, module_args = [], data = [], mounts = [], **kwargs):
    cli_wasm_run(
        name = name,
        module = module,
        module_args = module_args,
        data = data,
        mounts = mounts,
        runner = "@wasmedge//:bin/wasmedge",
        runner_files = ["@wasmedge//:runtime_files"],
        mount_flag = "--dir",
        mount_value = "{guest}:{host}",
        **kwargs
    )
