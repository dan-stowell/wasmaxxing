// Command wa is a minimal driver for the Wa programming language toolchain
// (wa-lang.org/wa), built on its public `api` package. Wa is a compiler that
// targets WebAssembly, and its compiler + a small embedded wazero runtime are
// pure Go — so this whole program cross-compiles to wasm (GOOS=wasip1) and runs
// on any WASI runtime. The result is a wasm-targeting compiler that itself runs
// inside WebAssembly: it compiles Wa source to wasm and executes it, all from
// within an outer runtime.
//
// Usage:
//
//	wa file.wa [args...]   # compile and run a Wa program (default)
//	wa -e 'func main { println(40 + 2) }'
//	wa -build file.wa      # print the compiled WebAssembly text (WAT)
//	wa                     # read Wa source from stdin
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"wa-lang.org/wa/api"
)

func main() {
	os.Exit(run())
}

func run() int {
	eval := flag.String("e", "", "compile and run a Wa code string instead of a file")
	build := flag.Bool("build", false, "print the compiled WebAssembly text (WAT) instead of running")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: wa [-build] [-e CODE] [file.wa] [args...]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	filename := "main.wa"
	var code string
	rest := flag.Args()
	switch {
	case *eval != "":
		code = *eval
	case flag.NArg() > 0:
		filename = flag.Arg(0)
		b, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wa: %v\n", err)
			return 1
		}
		code = string(b)
		rest = flag.Args()[1:]
	default:
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wa: reading stdin: %v\n", err)
			return 1
		}
		code = string(b)
	}

	cfg := api.DefaultConfig()

	// -build: emit the compiler's WebAssembly text output (the compiler
	// producing wasm from inside wasm), without running it.
	if *build {
		_, wat, _, err := api.BuildFile(cfg, filename, code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wa: %v\n", err)
			return 1
		}
		os.Stdout.Write(wat)
		return 0
	}

	// Default: compile and run, printing the program's combined output.
	out, err := api.RunCode(cfg, filename, code, rest...)
	os.Stdout.Write(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wa: %v\n", err)
		return 1
	}
	return 0
}
