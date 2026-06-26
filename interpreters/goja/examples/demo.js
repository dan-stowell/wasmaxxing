// Plain JavaScript, executed by the goja engine — which is itself running as a
// WebAssembly module on a WASI runtime.
console.log("hello from JavaScript, running on a Go engine compiled to wasm");

function fib(n) {
  var a = 0, b = 1;
  for (var i = 0; i < n; i++) { var t = a + b; a = b; b = t; }
  return a;
}

var out = [];
for (var i = 1; i <= 10; i++) out.push(fib(i));
console.log("first 10 Fibonacci numbers:", out.join(" "));

if (argv.length) console.log("argv:", argv.join(", "));
