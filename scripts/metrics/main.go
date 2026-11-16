// Command metrics records ShorStop's adoption numbers from the GitHub API into
// CSV files, one row per day, so trends survive GitHub's short retention (the
// traffic API keeps only 14 days). The metrics workflow runs it daily and
// commits the files to the "metrics" branch.
//
//	go run ./scripts/metrics -dir metrics-data
//
// GITHUB_TOKEN, if set, raises the API rate limit. TRAFFIC_TOKEN, if set, must
// be able to read the repository's traffic (a fine-grained token with
// "Administration: read-only" on the repository); without it the traffic
// files are left alone.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	repo := flag.String("repo", "shorstophq/shorstop", "owner/name of the repository")
	dir := flag.String("dir", "metrics-data", "directory holding the CSV files")
	api := flag.String("api", "https://api.github.com", "GitHub API base URL")
	flag.Parse()

	c := &client{
		base:         *api,
		repo:         *repo,
		token:        os.Getenv("GITHUB_TOKEN"),
		trafficToken: os.Getenv("TRAFFIC_TOKEN"),
	}
	today := time.Now().UTC().Format(time.DateOnly)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	err := run(ctx, c, *dir, today)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "metrics:", err)
		os.Exit(1)
	}
}
