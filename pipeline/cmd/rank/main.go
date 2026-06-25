// Command rank sorts catalog entries by the maturity heuristic and prints a
// table with the score breakdown.
//
//	bazel run //pipeline/cmd/rank -- -kind runtime
//	bazel run //pipeline/cmd/rank -- -kind runtime -top 10
//
// The maturity score blends adoption, activity, longevity, depth and
// governance; see package pipeline/maturity for the exact weights.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
	"github.com/dan-stowell/wasmaxxing/pipeline/maturity"
)

func workspaceRoot() string {
	if d := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); d != "" {
		return d
	}
	wd, _ := os.Getwd()
	return wd
}

func main() {
	in := flag.String("in", "data/catalog.json", "catalog path")
	kind := flag.String("kind", "runtime", "kind to rank (runtime, compiler, ...); empty = all")
	top := flag.Int("top", 0, "show only the top N (0 = all with data)")
	flag.Parse()

	inPath := *in
	if !filepath.IsAbs(inPath) {
		inPath = filepath.Join(workspaceRoot(), inPath)
	}
	c, err := catalog.LoadFile(inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		os.Exit(1)
	}

	entries := c.Entries
	if *kind != "" {
		entries = maturity.FilterByKind(entries, catalog.Kind(*kind))
	}
	ranked := maturity.Rank(entries)

	label := *kind
	if label == "" {
		label = "entries"
	}
	fmt.Printf("Maturity ranking of %ss (score = adoption.30 activity.25 longevity.15 depth.20 governance.10; archived \u00d70.25)\n\n", label)

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "#\tSCORE\tNAME\tSTARS\tCOMMITS\tCONTRIB\tLAST PUSH\tLIC\tREPO")
	rank := 0
	for _, s := range ranked {
		g := s.Entry.GitHub
		if g == nil || g.Error != "" {
			continue // skip un-assessable entries in the table
		}
		rank++
		if *top > 0 && rank > *top {
			break
		}
		commits := "-"
		contrib := "-"
		if s.HasDeep {
			commits = fmt.Sprintf("%d", g.Commits)
			contrib = fmt.Sprintf("%d", g.Contributors)
		}
		lic := g.License
		if lic == "" {
			lic = "-"
		}
		arch := ""
		if g.Archived {
			arch = " (archived)"
		}
		push := g.PushedAt
		if len(push) >= 10 {
			push = push[:10]
		}
		fmt.Fprintf(w, "%d\t%.3f\t%s%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			rank, s.Total, s.Entry.Name, arch, g.Stars, commits, contrib, push, lic, s.Entry.Repo)
	}
	w.Flush()
}
