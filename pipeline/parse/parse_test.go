package parse

import (
	"testing"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
)

const langsFixture = "### <a name=\"rust\"></a>Rust <sup>[top](#contents)</sup>\n" +
	"> Rust is a systems programming language.\n" +
	"* [wasm32 target](https://www.hellorust.com/setup/) - Rust backend for wasm.\n" +
	"* [Wasm-Bindgen](https://github.com/rustwasm/wasm-bindgen) - high-level interop.\n" +
	"\n--------------------\n" +
	"### <a name=\"go\"></a>Go <sup>[top](#contents)</sup>\n" +
	"> Go is an open source language.\n" +
	"* [TinyGo](https://github.com/tinygo-org/tinygo) - small wasm files.\n"

func TestParseLangs(t *testing.T) {
	es := Parse(FormatLangs, "langs.md", langsFixture)
	var langs, compilers int
	var bindgen *catalog.Entry
	for i := range es {
		switch es[i].Kind {
		case catalog.KindLanguage:
			langs++
		case catalog.KindCompiler:
			compilers++
			if es[i].Name == "Wasm-Bindgen" {
				bindgen = &es[i]
			}
		}
	}
	if langs != 2 {
		t.Errorf("languages=%d want 2", langs)
	}
	if compilers != 3 {
		t.Errorf("compilers=%d want 3", compilers)
	}
	if bindgen == nil {
		t.Fatal("missing Wasm-Bindgen")
	}
	if bindgen.RelatedLanguage != "Rust" {
		t.Errorf("related=%q want Rust", bindgen.RelatedLanguage)
	}
	if es[0].Description == "" {
		t.Error("rust language missing description")
	}
}

const runtimesFixture = "## <a name=\"wasmtime\"></a>[Wasmtime](https://github.com/CraneStation/wasmtime) <sup>[top](#c)</sup>\n" +
	"> Wasmtime is a standalone wasm-only runtime using Cranelift JIT\n" +
	"\n* **Languages written in**\n" +
	"## <a name=\"wasmer\"></a>[Wasmer](https://github.com/wasmerio/wasmer) <sup>[top](#c)</sup>\n" +
	"> Wasmer is a fast and secure WebAssembly runtime\n"

func TestParseRuntimes(t *testing.T) {
	es := Parse(FormatRuntimes, "rt.md", runtimesFixture)
	if len(es) != 2 {
		t.Fatalf("got %d want 2", len(es))
	}
	if es[0].Name != "Wasmtime" || es[0].Kind != catalog.KindRuntime {
		t.Errorf("entry0=%+v", es[0])
	}
	if es[0].Description == "" || es[0].URL == "" {
		t.Errorf("missing desc/url: %+v", es[0])
	}
}

const awesomeFixture = "## Compilers\n" +
	"- [Emscripten - LLVM-based C/C++ compiler](http://emscripten.org/)\n" +
	"- [Binaryen - toolchain library](https://github.com/WebAssembly/binaryen)\n" +
	"## Tools\n" +
	"### Kits\n" +
	"- [wabt - the WebAssembly Binary Toolkit](https://github.com/WebAssembly/wabt)\n"

func TestParseAwesome(t *testing.T) {
	es := Parse(FormatAwesome, "aw.md", awesomeFixture)
	if len(es) != 3 {
		t.Fatalf("got %d want 3", len(es))
	}
	if es[0].Name != "Emscripten" || es[0].Kind != catalog.KindCompiler {
		t.Errorf("entry0=%+v", es[0])
	}
	if es[0].Description == "" {
		t.Errorf("emscripten desc empty: %+v", es[0])
	}
	if es[2].Kind != catalog.KindTool {
		t.Errorf("wabt kind=%q want tool", es[2].Kind)
	}
}
