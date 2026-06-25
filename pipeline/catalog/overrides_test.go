package catalog

import "testing"

func TestApplyOverrides(t *testing.T) {
	c := Build([]Entry{
		{Name: "Wasm3", Kind: KindCompiler, URL: "https://github.com/wasm3/wasm3"},
		{Name: "wazero", Kind: KindRuntime, URL: "https://wazero.io"},
	})
	set := &OverrideSet{Overrides: []Override{
		{ID: "wasm3", Kind: KindRuntime},
		{ID: "wazero", Repo: "tetratelabs/wazero", URL: "https://github.com/tetratelabs/wazero"},
	}}
	c.Apply(set)

	byID := map[string]Entry{}
	for _, e := range c.Entries {
		byID[e.ID] = e
	}
	if byID["wasm3"].Kind != KindRuntime {
		t.Errorf("wasm3 kind=%q want runtime", byID["wasm3"].Kind)
	}
	if byID["wazero"].Repo != "tetratelabs/wazero" {
		t.Errorf("wazero repo=%q", byID["wazero"].Repo)
	}
	hasCurated := false
	for _, tag := range byID["wazero"].Tags {
		if tag == "curated" {
			hasCurated = true
		}
	}
	if !hasCurated {
		t.Error("expected 'curated' tag on overridden entry")
	}
}

func TestApplyAliasesMerge(t *testing.T) {
	c := Build([]Entry{
		{Name: "SSVM", Kind: KindRuntime, URL: "https://github.com/second-state/SSVM"},
		{Name: "WasmEdge", Kind: KindRuntime, URL: "https://github.com/WasmEdge/WasmEdge", Description: "the edge runtime"},
	})
	set := &OverrideSet{Overrides: []Override{
		{ID: "wasmedge", Aliases: []string{"ssvm"}},
	}}
	before := len(c.Entries)
	c.Apply(set)
	if len(c.Entries) != before-1 {
		t.Fatalf("alias not merged: have %d entries, want %d", len(c.Entries), before-1)
	}
	for _, e := range c.Entries {
		if e.ID == "ssvm" {
			t.Error("alias 'ssvm' should have been removed")
		}
	}
}

func TestMergeGitHubFrom(t *testing.T) {
	prev := &Catalog{Entries: []Entry{
		{ID: "wazero", Name: "wazero", Repo: "tetratelabs/wazero", GitHub: &GitHubInfo{Stars: 6000}},
	}}
	cur := Build([]Entry{
		{Name: "wazero", Kind: KindRuntime, URL: "https://github.com/tetratelabs/wazero"},
	})
	if n := cur.MergeGitHubFrom(prev); n != 1 {
		t.Fatalf("carried %d want 1", n)
	}
	if cur.Entries[0].GitHub == nil || cur.Entries[0].GitHub.Stars != 6000 {
		t.Errorf("github data not carried forward: %+v", cur.Entries[0].GitHub)
	}
}
