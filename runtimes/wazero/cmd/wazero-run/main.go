// Command wazero-run executes a WebAssembly module on the wazero runtime
// (a zero-dependency, pure-Go wasm runtime).
//
// It wires up WASI (wasip1) plus stdio and argv so that modules compiled from
// Go, Rust, C, etc. with a WASI target can run unmodified:
//
//	bazel run //runtimes/wazero/cmd/wazero-run -- path/to/module.wasm [args...]
//
// The first non-flag argument is the module path; any remaining arguments are
// passed to the module as its os.Args (argv[0] is derived from the filename).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

// mountFlag collects repeatable -dir HOST[:GUEST] filesystem mounts.
type mountFlag []string

func (m *mountFlag) String() string { return strings.Join(*m, ",") }
func (m *mountFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	os.Exit(run())
}

func run() int {
	var mounts mountFlag
	flag.Var(&mounts, "dir", "mount a host directory into the module as HOST[:GUEST] (repeatable)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: wazero-run [flags] MODULE.wasm [args...]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return 2
	}
	modPath := flag.Arg(0)
	modArgs := flag.Args()[1:]

	wasmBytes, err := os.ReadFile(modPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wazero-run: %v\n", err)
		return 1
	}

	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Assemble the guest filesystem from -dir mounts.
	fsConfig := wazero.NewFSConfig()
	for _, m := range mounts {
		host, guest := m, m
		if i := strings.IndexByte(m, ':'); i >= 0 {
			host, guest = m[:i], m[i+1:]
		}
		fsConfig = fsConfig.WithDirMount(host, guest)
	}

	// argv[0] is the module name; the rest are user-provided args.
	argv := append([]string{filepath.Base(modPath)}, modArgs...)
	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		WithArgs(argv...).
		WithSysWalltime().
		WithSysNanotime().
		WithSysNanosleep().
		WithRandSource(nil). // nil => crypto/rand
		WithFSConfig(fsConfig)

	// Inherit the host environment so modules can read env vars.
	for _, kv := range os.Environ() {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				config = config.WithEnv(kv[:i], kv[i+1:])
				break
			}
		}
	}

	_, err = r.InstantiateWithConfig(ctx, wasmBytes, config)
	if err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok {
			return int(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "wazero-run: %v\n", err)
		return 1
	}
	return 0
}
