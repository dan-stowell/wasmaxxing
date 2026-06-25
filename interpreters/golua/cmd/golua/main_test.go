package main

import (
	"testing"

	"github.com/Shopify/go-lua"
)

// TestSetArgTable verifies that arg[0] is the script name and arg[1..] are the
// remaining arguments, as Lua scripts expect.
func TestSetArgTable(t *testing.T) {
	l := lua.NewState()
	lua.OpenLibraries(l)
	setArgTable(l, "script.lua", []string{"one", "two"})

	if err := lua.DoString(l, `return arg[0], arg[1], arg[2]`); err != nil {
		t.Fatalf("DoString: %v", err)
	}
	got := make([]string, 3)
	for i := range got {
		s, ok := l.ToString(i - 3)
		if !ok {
			t.Fatalf("arg %d not a string", i)
		}
		got[i] = s
	}
	want := []string{"script.lua", "one", "two"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

// TestLuaSemantics is a guard against VM correctness regressions: the
// multiple-assignment swap that gopher-lua got wrong must work here.
func TestLuaSemantics(t *testing.T) {
	l := lua.NewState()
	lua.OpenLibraries(l)
	if err := lua.DoString(l, `
		local x, y = 5, 9
		x, y = y, x
		assert(x == 9 and y == 5, "swap failed")
		local a, b = 0, 1
		for _ = 1, 9 do a, b = b, a + b end
		assert(a == 34, "fib failed: " .. a)
	`); err != nil {
		t.Fatalf("semantics: %v", err)
	}
}
