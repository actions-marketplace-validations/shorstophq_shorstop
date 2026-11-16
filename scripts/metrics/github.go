package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// client reads the few GitHub API endpoints the metrics need.
type client struct {
	base, repo   string
	token        string // optional: raises the rate limit for public data
	trafficToken string // optional: needed for the traffic endpoints
	http         http.Client
}

type repoInfo struct {
	Stars       int `json:"stargazers_count"`
	Forks       int `json:"forks_count"`
	Watchers    int `json:"subscribers_count"` // "watchers_count" is stars, for historical reasons
	OpenIssues  int `json:"open_issues_count"` // issues and pull requests
	NetworkSize int `json:"network_count"`
}

type release struct {
	Tag        string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Draft      bool    `json:"draft"`
	Assets     []asset `json:"assets"`
}

type asset struct {
	Name      string `json:"name"`
	Downloads int    `json:"download_count"`
}

// trafficDay is one day of views or clones.
type trafficDay struct {
	Timestamp string `json:"timestamp"` // 2026-11-16T00:00:00Z
	Count     int    `json:"count"`
	Uniques   int    `json:"uniques"`
}

type referrer struct {
	Referrer string `json:"referrer"`
	Count    int    `json:"count"`
	Uniques  int    `json:"uniques"`
}

type popularPath struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Count   int    `json:"count"`
	Uniques int    `json:"uniques"`
}

func (c *client) repoInfo(ctx context.Context) (repoInfo, error) {
	var r repoInfo
	return r, c.get(ctx, "/repos/"+c.repo, c.token, &r)
}

// releases returns up to 100 releases, newest first. ShorStop won't have more
// for a long time; pagination can be added then.
func (c *client) releases(ctx context.Context) ([]release, error) {
	var rs []release
	return rs, c.get(ctx, "/repos/"+c.repo+"/releases?per_page=100", c.token, &rs)
}

func (c *client) views(ctx context.Context) ([]trafficDay, error) {
	var v struct {
		Views []trafficDay `json:"views"`
	}
	return v.Views, c.get(ctx, "/repos/"+c.repo+"/traffic/views?per=day", c.trafficToken, &v)
}

func (c *client) clones(ctx context.Context) ([]trafficDay, error) {
	var v struct {
		Clones []trafficDay `json:"clones"`
	}
	return v.Clones, c.get(ctx, "/repos/"+c.repo+"/traffic/clones?per=day", c.trafficToken, &v)
}

func (c *client) referrers(ctx context.Context) ([]referrer, error) {
	var rs []referrer
	return rs, c.get(ctx, "/repos/"+c.repo+"/traffic/popular/referrers", c.trafficToken, &rs)
}

func (c *client) paths(ctx context.Context) ([]popularPath, error) {
	var ps []popularPath
	return ps, c.get(ctx, "/repos/"+c.repo+"/traffic/popular/paths", c.trafficToken, &ps)
}

func (c *client) get(ctx context.Context, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("GET %s: %s: %s", path, resp.Status, msg)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
