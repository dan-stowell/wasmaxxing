// Command enrich-catalog augments data/catalog.json with GitHub repository
// metadata (stars, license, last push, archived status, ...).
//
// It requires network access and a GitHub token in GITHUB_TOKEN or GH_TOKEN.
//
//	GITHUB_TOKEN=$(gh auth token) bazel run //pipeline/cmd/enrich-catalog
//
// Flags:
//
//	-in/-out  catalog path (default data/catalog.json)
//	-limit    max repos to fetch (0 = all)
//	-only-missing  only fetch entries without existing GitHub data
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
	"github.com/dan-stowell/wasmaxxing/pipeline/github"
)

func workspaceRoot() string {
	if d := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); d != "" {
		return d
	}
	wd, _ := os.Getwd()
	return wd
}

func main() {
	in := flag.String("in", "data/catalog.json", "input catalog path")
	out := flag.String("out", "", "output path (defaults to -in)")
	limit := flag.Int("limit", 0, "max repos to fetch (0 = all)")
	onlyMissing := flag.Bool("only-missing", true, "only fetch entries lacking GitHub data")
	delay := flag.Duration("delay", 80*time.Millisecond, "delay between requests")
	flag.Parse()

	root := workspaceRoot()
	inPath := filepath.Join(root, *in)
	outPath := *out
	if outPath == "" {
		outPath = *in
	}
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(root, outPath)
	}

	c, err := catalog.LoadFile(inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		os.Exit(1)
	}

	client := github.NewClient()
	if client.Token == "" {
		fmt.Fprintln(os.Stderr, "warning: no GITHUB_TOKEN/GH_TOKEN set; low rate limit")
	}

	ctx := context.Background()
	fetched, skipped, failed := 0, 0, 0
	for i := range c.Entries {
		e := &c.Entries[i]
		if e.Repo == "" {
			continue
		}
		if *onlyMissing && e.GitHub != nil && e.GitHub.Error == "" {
			skipped++
			continue
		}
		if *limit > 0 && fetched >= *limit {
			break
		}
		info := client.Fetch(ctx, e.Repo)
		e.GitHub = info
		fetched++
		if info.Error != "" {
			failed++
			fmt.Fprintf(os.Stderr, "  %-40s ERR %s\n", e.Repo, info.Error)
		} else {
			fmt.Fprintf(os.Stderr, "  %-40s %5d\u2b50\n", e.Repo, info.Stars)
		}
		time.Sleep(*delay)
	}

	if err := c.WriteFile(outPath); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "\nenriched: %d fetched, %d skipped, %d failed -> %s\n",
		fetched, skipped, failed, outPath)
}
