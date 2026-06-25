// Command golua is a small Lua interpreter built on Shopify/go-lua, a pure-Go
// port of the Lua 5.2 reference VM. Being pure Go (no cgo), it compiles
// unmodified to wasm (GOOS=wasip1) and runs on any WASI runtime — see
// //interpreters/golua:golua_wasm and the wazero_run targets.
//
// Usage:
//
//	golua script.lua [args...]   # run a Lua script file
//	golua -e 'print(1+1)'        # run a Lua snippet
//	golua                        # read a script from stdin
//
// Command-line arguments are exposed to the script as the conventional global
// table `arg` (arg[0] = script name, arg[1..] = remaining arguments).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Shopify/go-lua"
)

func main() {
	os.Exit(run())
}

func run() int {
	eval := flag.String("e", "", "execute a Lua statement string instead of a file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: golua [-e CODE] [script.lua] [args...]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	l := lua.NewState()
	lua.OpenLibraries(l)

	// Determine script name and the trailing args for the `arg` table.
	scriptName := "=(stdin)"
	rest := flag.Args()
	if *eval == "" && len(rest) > 0 {
		scriptName = rest[0]
		rest = rest[1:]
	}
	setArgTable(l, scriptName, rest)

	var err error
	switch {
	case *eval != "":
		err = lua.DoString(l, *eval)
	case flag.NArg() > 0:
		err = lua.DoFile(l, flag.Arg(0))
	default:
		src, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "golua: reading stdin: %v\n", rerr)
			return 1
		}
		err = lua.DoString(l, string(src))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "golua: %v\n", err)
		return 1
	}
	return 0
}

// setArgTable builds the global `arg` table: arg[0]=name, arg[1..]=args.
func setArgTable(l *lua.State, name string, args []string) {
	l.NewTable()
	l.PushString(name)
	l.RawSetInt(-2, 0)
	for i, a := range args {
		l.PushString(a)
		l.RawSetInt(-2, i+1)
	}
	l.SetGlobal("arg")
}
