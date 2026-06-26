// AssemblyScript: a typed subset of TypeScript that compiles to WebAssembly.
// Compiled by `asc` (driven by a hermetic Node.js) into a WASI command module,
// this runs unmodified on every runtime in the repo.
console.log("hello from AssemblyScript, compiled to WebAssembly (WASI)");

// A little computation so there's something to see.
function fib(n: i32): i32 {
  let a = 0, b = 1;
  for (let i = 0; i < n; i++) {
    const t = a + b;
    a = b;
    b = t;
  }
  return a;
}

for (let i = 1; i <= 10; i++) {
  console.log("  fib(" + i.toString() + ") = " + fib(i).toString());
}
