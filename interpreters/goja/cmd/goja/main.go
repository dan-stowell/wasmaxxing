// Command goja is a small JavaScript (ECMAScript 5.1+) interpreter built on
// dop251/goja, a pure-Go ECMAScript engine. Being pure Go (no cgo), it compiles
// unmodified to wasm (GOOS=wasip1) and runs on any WASI runtime — so this is a
// JavaScript engine executing JavaScript from *inside* a WebAssembly module.
//
// Usage:
//
//	goja script.js [args...]    # run a JavaScript file
//	goja -e '1+1'               # evaluate a snippet
//	goja                        # read a script from stdin
//
// A minimal host environment is provided: console.log/info/warn/error and
// print() write to stdout; command-line arguments are exposed to the script as
// the global array `argv`.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dop251/goja"
)

func main() {
	os.Exit(run())
}

func run() int {
	eval := flag.String("e", "", "evaluate a JavaScript string instead of a file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: goja [-e CODE] [script.js] [args...]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	vm := goja.New()

	// Minimal console: join arguments with spaces, like browsers/Node.
	logTo := func(w io.Writer) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			parts := make([]string, len(call.Arguments))
			for i, a := range call.Arguments {
				parts[i] = a.String()
			}
			fmt.Fprintln(w, strings.Join(parts, " "))
			return goja.Undefined()
		}
	}
	out := logTo(os.Stdout)
	console := vm.NewObject()
	_ = console.Set("log", out)
	_ = console.Set("info", out)
	_ = console.Set("warn", logTo(os.Stderr))
	_ = console.Set("error", logTo(os.Stderr))
	_ = vm.Set("console", console)
	_ = vm.Set("print", out)

	// Expose trailing args as the global `argv`.
	name := "(stdin)"
	rest := flag.Args()
	if *eval == "" && len(rest) > 0 {
		name = rest[0]
		rest = rest[1:]
	}
	_ = vm.Set("argv", rest)

	var src string
	switch {
	case *eval != "":
		src, name = *eval, "(-e)"
	case flag.NArg() > 0:
		b, err := os.ReadFile(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "goja: %v\n", err)
			return 1
		}
		src = string(b)
	default:
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goja: reading stdin: %v\n", err)
			return 1
		}
		src = string(b)
	}

	if _, err := vm.RunScript(name, src); err != nil {
		fmt.Fprintf(os.Stderr, "goja: %v\n", err)
		return 1
	}
	return 0
}
