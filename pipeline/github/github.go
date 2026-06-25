// Package github fetches lightweight repository metadata from the GitHub REST
// API to enrich catalog entries.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	info.PushedAt = r.PushedAt
	info.DefaultBranch = r.DefaultBranch
	info.Homepage = r.Homepage
	info.Topics = r.Topics
	if r.License != nil && r.License.SPDX != "" && r.License.SPDX != "NOASSERTION" {
		info.License = r.License.SPDX
	}
	return info
}
