// Command build-catalog parses the committed seed markdown lists into a
// normalized data/catalog.json.
//
// Run it with Bazel so it can locate the workspace root:
//
//	bazel run //pipeline/cmd/build-catalog
//
// Flags:
//
//	-out   output path (default data/catalog.json, relative to workspace root)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
	"github.com/dan-stowell/wasmaxxing/pipeline/parse"
)

func workspaceRoot() string {
	if d := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); d != "" {
		return d
	}
	wd, _ := os.Getwd()
	return wd
}

func main() {
	out := flag.String("out", "data/catalog.json", "output path relative to workspace root")
	flag.Parse()

	root := workspaceRoot()
	var all []catalog.Entry
	for _, seed := range parse.DefaultSeeds() {
		p := filepath.Join(root, seed.Path)
		data, err := os.ReadFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", seed.Path, err)
			continue
		}
		entries := parse.Parse(seed.Format, seed.Name, string(data))
		fmt.Fprintf(os.Stderr, "%-22s %4d entries\n", seed.Name, len(entries))
		all = append(all, entries...)
	}

	c := catalog.Build(all)
	outPath := *out
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(root, outPath)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}
	if err := c.WriteFile(outPath); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}

	counts := c.Counts()
	kinds := make([]string, 0, len(counts))
	for k := range counts {
		kinds = append(kinds, string(k))
	}
	sort.Strings(kinds)
	fmt.Fprintf(os.Stderr, "\nwrote %s: %d entries\n", outPath, len(c.Entries))
	for _, k := range kinds {
		fmt.Fprintf(os.Stderr, "  %-10s %d\n", k, counts[catalog.Kind(k)])
	}
}
