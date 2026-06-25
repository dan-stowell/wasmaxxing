package catalog

import (
	"encoding/json"
	"io"
	"os"
)

// Override is a hand-curated correction to a catalog entry, applied after the
// seed lists are parsed. It fixes miscategorizations, supplies missing repos,
// and folds in duplicate ids the automatic dedupe misses.
type Override struct {
	ID              string   `json:"id"`
	Kind            Kind     `json:"kind,omitempty"`
	Repo            string   `json:"repo,omitempty"`
	URL             string   `json:"url,omitempty"`
	RelatedLanguage string   `json:"related_language,omitempty"`
	Description     string   `json:"description,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	// Aliases are other entry ids that should be merged into this one.
	Aliases []string `json:"aliases,omitempty"`
}

// OverrideSet is the on-disk overrides document.
type OverrideSet struct {
	Overrides []Override `json:"overrides"`
}

// LoadOverrides reads an overrides document from JSON.
func LoadOverrides(r io.Reader) (*OverrideSet, error) {
	var o OverrideSet
	if err := json.NewDecoder(r).Decode(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

// LoadOverridesFile reads overrides from path. A missing file is not an error
// (returns an empty set), so overrides are optional.
func LoadOverridesFile(p string) (*OverrideSet, error) {
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &OverrideSet{}, nil
		}
		return nil, err
	}
	defer f.Close()
	return LoadOverrides(f)
}

// Apply mutates the catalog in place: first folding aliases into their target
// id, then applying field corrections. Call after Build (entries have ids).
func (c *Catalog) Apply(set *OverrideSet) {
	if set == nil || len(set.Overrides) == 0 {
		return
	}
	index := map[string]int{}
	for i, e := range c.Entries {
		index[e.ID] = i
	}

	// Fold aliases into their target, then drop the alias entries.
	remove := map[int]bool{}
	for _, ov := range set.Overrides {
		tgt, ok := index[ov.ID]
		if !ok {
			continue
		}
		for _, alias := range ov.Aliases {
			if ai, ok := index[alias]; ok && ai != tgt {
				c.Entries[tgt].Merge(c.Entries[ai])
				remove[ai] = true
			}
		}
	}

	// Apply field corrections.
	for _, ov := range set.Overrides {
		i, ok := index[ov.ID]
		if !ok {
			continue
		}
		e := &c.Entries[i]
		if ov.Kind != "" {
			e.Kind = ov.Kind
		}
		if ov.Repo != "" {
			e.Repo = ov.Repo
		}
		if ov.URL != "" {
			e.URL = ov.URL
		}
		if ov.RelatedLanguage != "" {
			e.RelatedLanguage = ov.RelatedLanguage
		}
		if ov.Description != "" {
			e.Description = ov.Description
		}
		if len(ov.Tags) > 0 {
			e.Tags = mergeStrings(e.Tags, ov.Tags)
		}
		e.Tags = mergeStrings(e.Tags, []string{"curated"})
	}

	if len(remove) == 0 {
		return
	}
	kept := c.Entries[:0]
	for i, e := range c.Entries {
		if !remove[i] {
			kept = append(kept, e)
		}
	}
	c.Entries = kept
}
