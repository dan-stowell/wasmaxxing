// Package github fetches lightweight repository metadata from the GitHub REST
// API to enrich catalog entries.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/dan-stowell/wasmaxxing/pipeline/catalog"
)

// Client talks to the GitHub API. Token is optional but strongly recommended
// (raises rate limit from 60 to 5000 req/hr).
type Client struct {
	HTTP    *http.Client
	Token   string
	BaseURL string
}

// NewClient builds a Client, reading a token from GITHUB_TOKEN or GH_TOKEN.
func NewClient() *Client {
	tok := os.Getenv("GITHUB_TOKEN")
	if tok == "" {
		tok = os.Getenv("GH_TOKEN")
	}
	return &Client{
		HTTP:    &http.Client{Timeout: 20 * time.Second},
		Token:   tok,
		BaseURL: "https://api.github.com",
	}
}

type repoResponse struct {
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	OpenIssuesCount int    `json:"open_issues_count"`
	Language        string `json:"language"`
	Archived        bool   `json:"archived"`
	CreatedAt       string `json:"created_at"`
	PushedAt        string `json:"pushed_at"`
	DefaultBranch   string `json:"default_branch"`
	Homepage        string `json:"homepage"`
	Topics          []string `json:"topics"`
	License         *struct {
		SPDX string `json:"spdx_id"`
	} `json:"license"`
}

// Fetch retrieves metadata for "owner/name". Network/HTTP errors are returned
// inside GitHubInfo.Error rather than as a hard error, so a bad repo doesn't
// abort a batch.
func (c *Client) Fetch(ctx context.Context, repo string) *catalog.GitHubInfo {
	now := time.Now().UTC().Format(time.RFC3339)
	info := &catalog.GitHubInfo{FetchedAt: now}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/repos/"+repo, nil)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("http %d", resp.StatusCode)
		return info
	}
	var r repoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		info.Error = err.Error()
		return info
	}
	info.FullName = r.FullName
	info.Description = r.Description
	info.Stars = r.StargazersCount
	info.Forks = r.ForksCount
	info.OpenIssues = r.OpenIssuesCount
	info.Language = r.Language
	info.Archived = r.Archived
	info.CreatedAt = r.CreatedAt
	info.PushedAt = r.PushedAt
	info.DefaultBranch = r.DefaultBranch
	info.Homepage = r.Homepage
	info.Topics = r.Topics
	if r.License != nil && r.License.SPDX != "" && r.License.SPDX != "NOASSERTION" {
		info.License = r.License.SPDX
	}
	return info
}

// reLastPage extracts the page number from a Link header entry like:
//
//	<https://api.github.com/...&page=87>; rel="last"
var reLastPage = regexp.MustCompile(`[?&]page=(\d+)>;\s*rel="last"`)

// countViaPagination issues a per_page=1 request and reads the total item count
// from the Link header's rel="last" page number. Returns (count, ok). This is
// the standard cheap way to total commits/contributors/releases without
// downloading every page. If there is no Link header, there is at most one
// page; we then report the number of items in the single response.
func (c *Client) countViaPagination(ctx context.Context, path string) (int, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return 0, false
	}
	q := req.URL.Query()
	q.Set("per_page", "1")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, false
	}
	if m := reLastPage.FindStringSubmatch(resp.Header.Get("Link")); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n, true
		}
	}
	// No "last" link: a single page. Count items in the body.
	var items []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return 0, false
	}
	return len(items), true
}

// FetchDeep augments info in place with approximate commit, contributor and
// release totals. These cost one extra request each, so they are opt-in.
func (c *Client) FetchDeep(ctx context.Context, repo string, info *catalog.GitHubInfo) {
	if n, ok := c.countViaPagination(ctx, "/repos/"+repo+"/commits"); ok {
		info.Commits = n
	}
	if n, ok := c.countViaPagination(ctx, "/repos/"+repo+"/contributors"); ok {
		info.Contributors = n
	}
	if n, ok := c.countViaPagination(ctx, "/repos/"+repo+"/releases"); ok {
		info.Releases = n
	}
}
