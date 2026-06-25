// Command hello-go-wasm is a tiny program compiled to the wasip1/wasm target
// by rules_go. It exercises stdout, args and the WASI clock so that running it
// on a runtime demonstrates real host interaction, not just an exit code.
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("hello from Go, compiled to WebAssembly (wasip1)")
	fmt.Printf("args: %v\n", os.Args)
	fmt.Printf("unix time: %d\n", time.Now().Unix())
	name := "world"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	fmt.Printf("hello, %s!\n", name)
}
