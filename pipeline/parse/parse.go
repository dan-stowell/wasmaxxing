// Package parse turns the seed "awesome" markdown lists into catalog entries.
//
// Three on-disk formats are supported, each with its own parser:
//
//   - FormatLangs:    appcypher/awesome-wasm-langs & wasmlang.org fork.
//                     "### <a name=ID></a>NAME" sections, each a language,
//                     with "* [Title](url) - desc" bullets for its toolchain.
//   - FormatRuntimes: appcypher/awesome-wasm-runtimes.
//                     "## <a name=ID></a>[NAME](url)" sections, each a runtime.
//   - FormatAwesome:  mbasso/awesome-wasm.
//                     "## Section" headings with "- [Title - desc](url)" bullets.
package parse

import (
	"regexp"
	"strings"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
)

// Format identifies a seed file layout.
type Format string

const (
	FormatLangs    Format = "langs"
	FormatRuntimes Format = "runtimes"
	FormatAwesome  Format = "awesome"
)

var (
	// ### <a name="rust"></a>Rust <sup>...   (langs detail headers)
	reLangHeader = regexp.MustCompile(`^###\s+<a name="([^"]+)"></a>\s*(.+?)\s*(<sup>.*)?$`)
	// ## <a name="wasmtime"></a>[Wasmtime](https://...) <sup>...  (runtime headers)
	reRuntimeHeader = regexp.MustCompile(`^##\s+<a name="([^"]+)"></a>\s*\[([^\]]+)\]\(([^)]+)\)`)
	// * [Title](url) - description   (langs bullets)
	reLinkBullet = regexp.MustCompile(`^[*-]\s*\[([^\]]+)\]\(([^)]+)\)\s*(?:[-—]\s*(.*))?$`)
	// ## Heading  /  ### Subheading
	reH2 = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	reH3 = regexp.MustCompile(`^###\s+(.+?)\s*$`)
)

// cleanText strips trailing markdown/html cruft and links from a fragment.
func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "</br>")
	s = strings.ReplaceAll(s, "</br>", "")
	s = strings.ReplaceAll(s, "<br>", "")
	// Drop trailing "<sup>..." navigation.
	if i := strings.Index(s, "<sup>"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func firstURL(s string) string {
	if m := regexp.MustCompile(`\(([^)]+)\)`).FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}

// Parse dispatches to the right format parser. sourceName is recorded on each
// entry (e.g. "wasm-langs.md").
func Parse(format Format, sourceName, content string) []catalog.Entry {
	switch format {
	case FormatLangs:
		return parseLangs(sourceName, content)
	case FormatRuntimes:
		return parseRuntimes(sourceName, content)
	case FormatAwesome:
		return parseAwesome(sourceName, content)
	default:
		return nil
	}
}

// parseLangs handles the awesome-wasm-langs / wasmlang.org format.
func parseLangs(source, content string) []catalog.Entry {
	var entries []catalog.Entry
	var curLang string
	inDetail := false

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, " \t")
		if m := reLangHeader.FindStringSubmatch(line); m != nil {
			inDetail = true
			name := cleanText(m[2])
			// Strip strikethrough markers used for unmaintained entries.
			name = strings.Trim(name, "~ ")
			curLang = name
			entries = append(entries, catalog.Entry{
				Name:     name,
				Kind:     catalog.KindLanguage,
				Category: name,
				Sources:  []string{source},
			})
			continue
		}
		if !inDetail {
			continue
		}
		// Description line for the current language.
		if strings.HasPrefix(line, ">") {
			desc := cleanText(strings.TrimPrefix(line, ">"))
			if n := len(entries); n > 0 && entries[n-1].Kind == catalog.KindLanguage && entries[n-1].Description == "" {
				entries[n-1].Description = desc
			}
			continue
		}
		// Toolchain bullet under a language: treat as a compiler.
		if m := reLinkBullet.FindStringSubmatch(line); m != nil && curLang != "" {
			title := cleanText(m[1])
			url := strings.TrimSpace(m[2])
			desc := cleanText(m[3])
			entries = append(entries, catalog.Entry{
				Name:            title,
				Kind:            catalog.KindCompiler,
				Description:     desc,
				URL:             url,
				Category:        curLang,
				RelatedLanguage: curLang,
				Sources:         []string{source},
			})
		}
	}
	return entries
}

// parseRuntimes handles the awesome-wasm-runtimes format.
func parseRuntimes(source, content string) []catalog.Entry {
	var entries []catalog.Entry
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		m := reRuntimeHeader.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := cleanText(m[2])
		url := strings.TrimSpace(m[3])
		// Look ahead for the blockquote description.
		desc := ""
		for j := i + 1; j < len(lines) && j < i+6; j++ {
			t := strings.TrimSpace(lines[j])
			if strings.HasPrefix(t, ">") {
				desc = cleanText(strings.TrimPrefix(t, ">"))
				break
			}
		}
		entries = append(entries, catalog.Entry{
			Name:        name,
			Kind:        catalog.KindRuntime,
			Description: desc,
			URL:         url,
			Category:    "runtime",
			Sources:     []string{source},
		})
	}
	return entries
}

// awesomeSectionKind maps a top-level awesome-wasm heading to an entry kind.
func awesomeSectionKind(section string) catalog.Kind {
	s := strings.ToLower(section)
	switch {
	case strings.Contains(s, "compiler"):
		return catalog.KindCompiler
	case strings.Contains(s, "embedding"), strings.Contains(s, "runtime"):
		return catalog.KindRuntime
	case strings.Contains(s, "tool"), strings.Contains(s, "editor"), strings.Contains(s, "kit"):
		return catalog.KindTool
	case strings.Contains(s, "language"), strings.Contains(s, "esoteric"):
		return catalog.KindLanguage
	case strings.Contains(s, "project"), strings.Contains(s, "framework"),
		strings.Contains(s, "demo"), strings.Contains(s, "example"):
		return catalog.KindProject
	default:
		return catalog.KindResource
	}
}

// parseAwesome handles the mbasso/awesome-wasm format.
func parseAwesome(source, content string) []catalog.Entry {
	var entries []catalog.Entry
	var topSection, subSection string
	kind := catalog.KindResource

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if m := reH3.FindStringSubmatch(line); m != nil {
			subSection = cleanText(m[1])
			// Sub-section kind may refine the top section (e.g. under Projects).
			if topSection != "" {
				kind = awesomeSectionKind(topSection)
			}
			continue
		}
		if m := reH2.FindStringSubmatch(line); m != nil {
			topSection = cleanText(m[1])
			subSection = ""
			kind = awesomeSectionKind(topSection)
			continue
		}
		// Only collect entries from sections that map to a concrete kind.
		if !strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "* ") {
			continue
		}
		m := reLinkBullet.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		inner := cleanText(m[1])
		url := strings.TrimSpace(m[2])
		desc := cleanText(m[3])
		// Awesome-wasm often packs "Name - description" inside the link text.
		name := inner
		if idx := strings.Index(inner, " - "); idx >= 0 {
			name = strings.TrimSpace(inner[:idx])
			if desc == "" {
				desc = strings.TrimSpace(inner[idx+3:])
			}
		}
		cat := topSection
		if subSection != "" {
			cat = subSection
		}
		entries = append(entries, catalog.Entry{
			Name:        name,
			Kind:        kind,
			Description: desc,
			URL:         url,
			Category:    cat,
			Sources:     []string{source},
		})
	}
	return entries
}

// SeedFile describes one seed input on disk.
type SeedFile struct {
	Path   string
	Name   string
	Format Format
}

// DefaultSeeds enumerates the committed seed files relative to the repo root.
func DefaultSeeds() []SeedFile {
	return []SeedFile{
		{Path: "data/sources/wasm-langs.md", Name: "awesome-wasm-langs", Format: FormatLangs},
		{Path: "data/sources/wasmlang-org.md", Name: "wasmlang.org", Format: FormatLangs},
		{Path: "data/sources/wasm-runtimes.md", Name: "awesome-wasm-runtimes", Format: FormatRuntimes},
		{Path: "data/sources/awesome-wasm.md", Name: "awesome-wasm", Format: FormatAwesome},
	}
}
