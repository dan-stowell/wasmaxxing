package maturity

import (
	"testing"
	"time"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
)

var testNow = time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)

func gh(mod func(*catalog.GitHubInfo)) *catalog.GitHubInfo {
	g := &catalog.GitHubInfo{
		Stars: 1000, Forks: 100, License: "Apache-2.0",
		CreatedAt: "2020-01-01T00:00:00Z",
		PushedAt:  "2026-06-20T00:00:00Z",
	}
	if mod != nil {
		mod(g)
	}
	return g
}

func TestArchivedIsPenalized(t *testing.T) {
	active := evaluateAt(catalog.Entry{GitHub: gh(nil)}, scoreOptions{now: testNow})
	archived := evaluateAt(catalog.Entry{GitHub: gh(func(g *catalog.GitHubInfo) { g.Archived = true })}, scoreOptions{now: testNow})
	if archived.Total >= active.Total {
		t.Errorf("archived (%.3f) should score below active (%.3f)", archived.Total, active.Total)
	}
}

func TestActivityDecays(t *testing.T) {
	fresh := evaluateAt(catalog.Entry{GitHub: gh(func(g *catalog.GitHubInfo) { g.PushedAt = "2026-06-24T00:00:00Z" })}, scoreOptions{now: testNow})
	stale := evaluateAt(catalog.Entry{GitHub: gh(func(g *catalog.GitHubInfo) { g.PushedAt = "2021-01-01T00:00:00Z" })}, scoreOptions{now: testNow})
	if fresh.Activity <= stale.Activity {
		t.Errorf("fresh activity %.3f should exceed stale %.3f", fresh.Activity, stale.Activity)
	}
	if fresh.Activity < 0.9 {
		t.Errorf("recently-pushed activity should be near 1, got %.3f", fresh.Activity)
	}
}

func TestAdoptionMonotonic(t *testing.T) {
	low := evaluateAt(catalog.Entry{GitHub: gh(func(g *catalog.GitHubInfo) { g.Stars = 50 })}, scoreOptions{now: testNow})
	high := evaluateAt(catalog.Entry{GitHub: gh(func(g *catalog.GitHubInfo) { g.Stars = 18000 })}, scoreOptions{now: testNow})
	if high.Adoption <= low.Adoption {
		t.Errorf("more stars should raise adoption: low=%.3f high=%.3f", low.Adoption, high.Adoption)
	}
}

func TestDeepMetricsRaiseDepth(t *testing.T) {
	shallow := evaluateAt(catalog.Entry{GitHub: gh(nil)}, scoreOptions{now: testNow})
	deep := evaluateAt(catalog.Entry{GitHub: gh(func(g *catalog.GitHubInfo) {
		g.Commits = 4000
		g.Contributors = 150
		g.Releases = 60
	})}, scoreOptions{now: testNow})
	if !deep.HasDeep {
		t.Error("expected HasDeep=true")
	}
	if deep.Depth <= shallow.Depth {
		t.Errorf("deep depth %.3f should exceed shallow %.3f", deep.Depth, shallow.Depth)
	}
}

func TestNoGitHubScoresZero(t *testing.T) {
	s := evaluateAt(catalog.Entry{Name: "x"}, scoreOptions{now: testNow})
	if s.Total != 0 {
		t.Errorf("entry without GitHub should score 0, got %.3f", s.Total)
	}
}

func TestRankOrders(t *testing.T) {
	entries := []catalog.Entry{
		{Name: "obscure", GitHub: gh(func(g *catalog.GitHubInfo) { g.Stars = 5; g.Forks = 1; g.License = "" })},
		{Name: "mature", GitHub: gh(func(g *catalog.GitHubInfo) { g.Stars = 18000; g.Forks = 1500; g.Commits = 9000; g.Contributors = 300; g.Releases = 80 })},
		{Name: "none"},
	}
	ranked := Rank(entries)
	if ranked[0].Entry.Name != "mature" {
		t.Errorf("expected 'mature' first, got %q", ranked[0].Entry.Name)
	}
	if ranked[len(ranked)-1].Entry.Name != "none" {
		t.Errorf("expected 'none' last, got %q", ranked[len(ranked)-1].Entry.Name)
	}
}
