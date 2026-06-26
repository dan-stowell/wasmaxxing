"""`wamr_run`: run a .wasm module on WAMR's `iwasm` as a `bazel run` target.

    load("//runtimes/wamr:defs.bzl", "wamr_run")

    wamr_run(
        name = "run_wamr",
        module = ":hello_wasm",
        module_args = ["world"],
    )

iwasm spells `iwasm [--map-dir=GUEST::HOST] MODULE [args...]`. We default to
`--interp` (the classic interpreter): the prebuilt's fast-JIT exhausts its code
cache on larger modules (e.g. the Lua interpreter), while the interpreter runs
everything reliably. We also raise `--stack-size` to 8 MB, since iwasm's 64 KB
default overflows on Go's deep wasm call stacks.
"""

load("//runtimes:cli_run.bzl", "cli_wasm_run")

def wamr_run(name, module, module_args = [], data = [], mounts = [], **kwargs):
    cli_wasm_run(
        name = name,
        module = module,
        module_args = module_args,
        data = data,
        mounts = mounts,
        runner = "@wamr//:iwasm",
        runner_files = ["@wamr//:runtime_files"],
        run_prefix = ["--interp", "--stack-size=8388608"],
        mount_flag = "--map-dir=",
        mount_value = "{guest}::{host}",
        mount_inline = True,
        **kwargs
    )
