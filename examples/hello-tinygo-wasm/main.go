// Command hello-tinygo-wasm is compiled to wasm by the TinyGo toolchain,
// demonstrating TinyGo's much smaller module size compared to the standard Go
// compiler (compare with //examples/hello-go-wasm).
package main

import "fmt"

func main() {
	fmt.Println("hello from TinyGo, compiled to WebAssembly (wasip1)")
	for i := 1; i <= 5; i++ {
		fmt.Printf("  tick %d\n", i)
	}
}
