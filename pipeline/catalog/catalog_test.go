package catalog

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Rust":          "rust",
		"C++":           "cpp",
		"C#":            "csharp",
		"Kotlin/Wasm":   "kotlin-wasm",
		"  Hello World": "hello-world",
		".Net":          "net",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q)=%q want %q", in, got, want)
		}
	}
}

func TestRepoFromURL(t *testing.T) {
	cases := map[string]string{
		"https://github.com/bytecodealliance/wasmtime":      "bytecodealliance/wasmtime",
		"https://github.com/CraneStation/wasmtime/":         "CraneStation/wasmtime",
		"https://github.com/foo/bar.git":                    "foo/bar",
		"https://github.com/foo/bar/tree/master/sub":        "foo/bar",
		"https://example.com/foo/bar":                       "",
		"https://github.com/topics/wasm":                    "",
		"not a url at all":                                  "",
	}
	for in, want := range cases {
		if got := RepoFromURL(in); got != want {
			t.Errorf("RepoFromURL(%q)=%q want %q", in, got, want)
		}
	}
}

func TestBuildDedupesByRepo(t *testing.T) {
	in := []Entry{
		{Name: "Wasmtime", Kind: KindRuntime, URL: "https://github.com/CraneStation/wasmtime", Sources: []string{"a"}},
		{Name: "Wasmtime", Kind: KindRuntime, URL: "https://github.com/CraneStation/wasmtime", Description: "jit", Sources: []string{"b"}},
	}
	c := Build(in)
	if len(c.Entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(c.Entries))
	}
	e := c.Entries[0]
	if e.Description != "jit" {
		t.Errorf("description not merged: %q", e.Description)
	}
	if len(e.Sources) != 2 {
		t.Errorf("sources not merged: %v", e.Sources)
	}
	if e.Repo != "CraneStation/wasmtime" {
		t.Errorf("repo=%q", e.Repo)
	}
}

func TestCounts(t *testing.T) {
	c := Build([]Entry{
		{Name: "A", Kind: KindRuntime, URL: "https://x/1"},
		{Name: "B", Kind: KindRuntime, URL: "https://x/2"},
		{Name: "C", Kind: KindCompiler, URL: "https://x/3"},
	})
	counts := c.Counts()
	if counts[KindRuntime] != 2 || counts[KindCompiler] != 1 {
		t.Errorf("counts=%v", counts)
	}
}
