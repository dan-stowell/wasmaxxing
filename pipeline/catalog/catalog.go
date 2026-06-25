// Package catalog defines the normalized data model for the wasmaxxing
// ecosystem catalog and helpers for reading/writing it as JSON.
//
// The catalog is the structured output of the data pipeline: a deduplicated
// list of WebAssembly languages, compilers, interpreters, runtimes, tools and
// projects gathered from the seed "awesome" lists and enriched with GitHub
// metadata.
package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
)

// Kind classifies an entry in the catalog.
type Kind string

const (
	KindLanguage Kind = "language" // a source language that compiles to / has a VM in wasm
	KindCompiler Kind = "compiler" // a tool that emits wasm
	KindRuntime  Kind = "runtime"  // an engine that executes wasm modules
	KindTool     Kind = "tool"     // a tool that operates on wasm (inspect, optimize, convert)
	KindProject  Kind = "project"  // an application or library built with/on wasm
	KindResource Kind = "resource" // docs, tutorials, articles, playgrounds, etc.
)

// Entry is a single normalized item in the catalog.
type Entry struct {
	// ID is a stable slug derived from Name (lowercase, hyphenated).
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind Kind   `json:"kind"`

	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`

	// Repo is "owner/name" when URL points at a GitHub repository.
	Repo string `json:"repo,omitempty"`

	// Category is the originating section/heading (e.g. "Rust", "Compilers").
	Category string `json:"category,omitempty"`

	// RelatedLanguage links a compiler/tool to the language section it appeared under.
	RelatedLanguage string `json:"related_language,omitempty"`

	// Sources lists the seed files this entry was found in.
	Sources []string `json:"sources,omitempty"`

	// Tags are free-form labels (e.g. "active", "unmaintained").
	Tags []string `json:"tags,omitempty"`

	// GitHub enrichment (populated by the enrich step). Nil until enriched.
	GitHub *GitHubInfo `json:"github,omitempty"`
}

// GitHubInfo holds metadata fetched from the GitHub API for an Entry's Repo.
type GitHubInfo struct {
	FullName      string `json:"full_name"`
	Description   string `json:"description,omitempty"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	OpenIssues    int    `json:"open_issues"`
	Language      string `json:"language,omitempty"`
	License       string `json:"license,omitempty"`
	Archived      bool   `json:"archived"`
	CreatedAt     string `json:"created_at,omitempty"`
	PushedAt      string `json:"pushed_at,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Homepage      string `json:"homepage,omitempty"`
	Topics        []string `json:"topics,omitempty"`
	FetchedAt     string `json:"fetched_at,omitempty"`

	// Deep metrics (populated only by enrich -deep). These are approximate
	// totals derived from the GitHub API's paginated Link headers.
	Commits      int `json:"commits,omitempty"`
	Contributors int `json:"contributors,omitempty"`
	Releases     int `json:"releases,omitempty"`

	// Error records a fetch failure (e.g. 404, rate limited) without aborting.
	Error string `json:"error,omitempty"`
}

// Catalog is the top-level document.
type Catalog struct {
	GeneratedAt string  `json:"generated_at,omitempty"`
	Entries     []Entry `json:"entries"`
}

// Slugify produces a stable lowercase hyphenated id from a name.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case r == '+':
			b.WriteString("p")
			lastHyphen = false
		case r == '#':
			b.WriteString("sharp")
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// RepoFromURL extracts "owner/name" if u is a GitHub repository URL, else "".
func RepoFromURL(u string) string {
	parsed, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Host)
	if host != "github.com" && host != "www.github.com" {
		return ""
	}
	segs := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segs) < 2 {
		return ""
	}
	owner, name := segs[0], segs[1]
	// Skip non-repo namespaces.
	switch strings.ToLower(owner) {
	case "", "sponsors", "orgs", "topics", "about", "features", "marketplace":
		return ""
	}
	name = strings.TrimSuffix(name, ".git")
	if name == "" {
		return ""
	}
	return owner + "/" + name
}

// dedupeKey identifies entries that should be merged.
func (e Entry) dedupeKey() string {
	if e.Repo != "" {
		return "repo:" + strings.ToLower(e.Repo)
	}
	if e.URL != "" {
		if p, err := url.Parse(e.URL); err == nil {
			return "url:" + strings.ToLower(p.Host) + path.Clean("/"+strings.Trim(p.Path, "/"))
		}
	}
	return "name:" + string(e.Kind) + ":" + e.ID
}

func mergeStrings(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range append(append([]string{}, a...), b...) {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// Merge combines a new entry into an existing one, preferring richer data.
func (e *Entry) Merge(o Entry) {
	if e.Description == "" {
		e.Description = o.Description
	}
	if e.URL == "" {
		e.URL = o.URL
	}
	if e.Repo == "" {
		e.Repo = o.Repo
	}
	if e.Category == "" {
		e.Category = o.Category
	}
	if e.RelatedLanguage == "" {
		e.RelatedLanguage = o.RelatedLanguage
	}
	e.Sources = mergeStrings(e.Sources, o.Sources)
	e.Tags = mergeStrings(e.Tags, o.Tags)
}

// Normalize fills derived fields (ID, Repo) for a single entry.
func (e *Entry) Normalize() {
	if e.ID == "" {
		e.ID = Slugify(e.Name)
	}
	if e.Repo == "" && e.URL != "" {
		e.Repo = RepoFromURL(e.URL)
	}
}

// Build normalizes, dedupes and sorts a set of entries into a Catalog.
func Build(entries []Entry) *Catalog {
	index := map[string]int{}
	var out []Entry
	for _, e := range entries {
		e.Normalize()
		if e.Name == "" {
			continue
		}
		key := e.dedupeKey()
		if i, ok := index[key]; ok {
			out[i].Merge(e)
			continue
		}
		index[key] = len(out)
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return &Catalog{Entries: out}
}

// MergeGitHubFrom carries forward GitHub enrichment from a previous catalog
// for entries with a matching repo (case-insensitive). Returns the number of
// entries that received preserved data. This lets build-catalog re-run without
// discarding the (expensive) enrich step's results.
func (c *Catalog) MergeGitHubFrom(prev *Catalog) int {
	if prev == nil {
		return 0
	}
	byRepo := map[string]*GitHubInfo{}
	for i := range prev.Entries {
		e := &prev.Entries[i]
		if e.Repo != "" && e.GitHub != nil {
			byRepo[strings.ToLower(e.Repo)] = e.GitHub
		}
	}
	n := 0
	for i := range c.Entries {
		e := &c.Entries[i]
		if e.GitHub != nil || e.Repo == "" {
			continue
		}
		if g, ok := byRepo[strings.ToLower(e.Repo)]; ok {
			e.GitHub = g
			n++
		}
	}
	return n
}

// Counts returns a per-kind tally.
func (c *Catalog) Counts() map[Kind]int {
	m := map[Kind]int{}
	for _, e := range c.Entries {
		m[e.Kind]++
	}
	return m
}

// Write serializes the catalog as pretty JSON.
func (c *Catalog) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(c)
}

// WriteFile writes the catalog to path.
func (c *Catalog) WriteFile(p string) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := c.Write(f); err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}

// Load reads a catalog from JSON.
func Load(r io.Reader) (*Catalog, error) {
	var c Catalog
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// LoadFile reads a catalog from path.
func LoadFile(p string) (*Catalog, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Load(f)
}
