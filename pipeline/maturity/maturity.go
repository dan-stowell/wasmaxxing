// Package maturity scores catalog entries by a transparent maturity heuristic.
//
// "Maturity" here is a blend of adoption, activity, longevity, depth of
// development, and governance signals derived from GitHub metadata. No single
// number captures project health, so the score is a weighted sum of normalized
// sub-scores, each clamped to [0,1], and the breakdown is exposed so callers
// can see *why* something ranked where it did.
//
// Signals and rationale:
//
//   - Adoption    (stars, forks): popularity / real-world use.
//   - Activity    (days since last push): is it still maintained?
//   - Longevity   (age in years): proven, not a flash in the pan.
//   - Depth       (commits, contributors, releases): sustained, multi-person
//                 effort with shipped versions. Requires `enrich -deep`.
//   - Governance  (has an OSI license): safe to depend on.
//
// Archived repositories are heavily penalized: they are explicitly
// end-of-life regardless of past popularity.
package maturity

import (
	"math"
	"sort"
	"time"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
)

// Weights for each sub-score. They sum to 1.0 so the final Score is in [0,1].
var (
	weightAdoption   = 0.30
	weightActivity   = 0.25
	weightLongevity  = 0.15
	weightDepth      = 0.20
	weightGovernance = 0.10
)

// Score is a maturity assessment with a breakdown of contributing factors.
type Score struct {
	Entry catalog.Entry

	// Total is the weighted overall score in [0,1].
	Total float64

	// Sub-scores, each in [0,1].
	Adoption   float64
	Activity   float64
	Longevity  float64
	Depth      float64
	Governance float64

	// HasDeep reports whether deep metrics (commits/contributors/releases)
	// were available; without them Depth is estimated from forks alone.
	HasDeep bool
}

// log1pNorm maps a non-negative count onto [0,1] via log scaling against a
// reference "full marks" value. Diminishing returns suit star/commit counts.
func log1pNorm(v float64, full float64) float64 {
	if v <= 0 {
		return 0
	}
	s := math.Log1p(v) / math.Log1p(full)
	return clamp01(s)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func parseTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// scoreOptions allows tests to pin "now" for deterministic results.
type scoreOptions struct {
	now time.Time
}

// Evaluate computes a maturity Score for one entry.
func Evaluate(e catalog.Entry) Score {
	return evaluateAt(e, scoreOptions{now: time.Now().UTC()})
}

func evaluateAt(e catalog.Entry, opt scoreOptions) Score {
	s := Score{Entry: e}
	g := e.GitHub
	if g == nil || g.Error != "" {
		return s // all zeros: no data => not assessable
	}

	// Adoption: stars dominate, forks add a smaller signal.
	stars := log1pNorm(float64(g.Stars), 20000)
	forks := log1pNorm(float64(g.Forks), 2000)
	s.Adoption = clamp01(0.75*stars + 0.25*forks)

	// Activity: decays with time since last push. <30d ~full, ~1yr ~0.5,
	// >3yr ~0. Uses an exponential half-life of one year.
	if pushed, ok := parseTime(g.PushedAt); ok {
		days := opt.now.Sub(pushed).Hours() / 24
		if days < 0 {
			days = 0
		}
		s.Activity = clamp01(math.Pow(0.5, days/365.0))
	}

	// Longevity: older projects are more proven. ~0 at birth, ~full by 6yr.
	if created, ok := parseTime(g.CreatedAt); ok {
		years := opt.now.Sub(created).Hours() / 24 / 365
		s.Longevity = log1pNorm(years, 6)
	}

	// Depth: sustained, multi-person, versioned development.
	if g.Commits > 0 || g.Contributors > 0 || g.Releases > 0 {
		s.HasDeep = true
		commits := log1pNorm(float64(g.Commits), 5000)
		contrib := log1pNorm(float64(g.Contributors), 200)
		releases := log1pNorm(float64(g.Releases), 100)
		s.Depth = clamp01(0.4*commits + 0.4*contrib + 0.2*releases)
	} else {
		// Fallback when deep metrics are absent: approximate from forks.
		s.Depth = log1pNorm(float64(g.Forks), 2000)
	}

	// Governance: presence of an OSI license id.
	if g.License != "" {
		s.Governance = 1
	}

	s.Total = weightAdoption*s.Adoption +
		weightActivity*s.Activity +
		weightLongevity*s.Longevity +
		weightDepth*s.Depth +
		weightGovernance*s.Governance

	// Archived projects are end-of-life: cap the score hard.
	if g.Archived {
		s.Total *= 0.25
	}
	return s
}

// Rank evaluates and sorts entries by descending maturity. Entries without
// GitHub data sink to the bottom (score 0).
func Rank(entries []catalog.Entry) []Score {
	scores := make([]Score, 0, len(entries))
	for _, e := range entries {
		scores = append(scores, Evaluate(e))
	}
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].Total != scores[j].Total {
			return scores[i].Total > scores[j].Total
		}
		return scores[i].Entry.GitHub != nil && scores[j].Entry.GitHub != nil &&
			scores[i].Entry.GitHub.Stars > scores[j].Entry.GitHub.Stars
	})
	return scores
}

// FilterByKind returns only entries of the given kind.
func FilterByKind(entries []catalog.Entry, kind catalog.Kind) []catalog.Entry {
	var out []catalog.Entry
	for _, e := range entries {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}
