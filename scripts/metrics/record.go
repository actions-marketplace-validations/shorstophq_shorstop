package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// The CSV files, each with a date as its first column. Rows for a date are
// replaced, never duplicated, so the job can run twice in a day.
var (
	repoCols      = []string{"date", "stars", "forks", "watchers", "open_issues"}
	downloadCols  = []string{"date", "release", "prerelease", "kind", "platform", "downloads"}
	viewCols      = []string{"date", "views", "unique_visitors"}
	cloneCols     = []string{"date", "clones", "unique_cloners"}
	referrerCols  = []string{"date", "referrer", "views_14d", "unique_visitors_14d"}
	popularPathCo = []string{"date", "path", "title", "views_14d", "unique_visitors_14d"}
)

func run(ctx context.Context, c *client, dir, today string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	info, err := c.repoInfo(ctx)
	if err != nil {
		return err
	}
	if err := upsert(filepath.Join(dir, "repo.csv"), repoCols, [][]string{{
		today, itoa(info.Stars), itoa(info.Forks), itoa(info.Watchers), itoa(info.OpenIssues),
	}}); err != nil {
		return err
	}

	rels, err := c.releases(ctx)
	if err != nil {
		return err
	}
	if err := upsert(filepath.Join(dir, "downloads.csv"), downloadCols, downloadRows(today, rels)); err != nil {
		return err
	}

	if c.trafficToken == "" {
		fmt.Fprintln(os.Stderr, "metrics: TRAFFIC_TOKEN not set; views, clones, referrers and paths not recorded")
	} else if err := recordTraffic(ctx, c, dir, today); err != nil {
		return err
	}

	summary, err := summarize(dir, c.repo)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(summary), 0o644)
}

// recordTraffic stores the traffic API's daily views and clones (each run
// refreshes the last 14 days, which the API may still be revising) and a
// snapshot of the top referrers and pages.
func recordTraffic(ctx context.Context, c *client, dir, today string) error {
	views, err := c.views(ctx)
	if err != nil {
		return err
	}
	clones, err := c.clones(ctx)
	if err != nil {
		return err
	}
	refs, err := c.referrers(ctx)
	if err != nil {
		return err
	}
	paths, err := c.paths(ctx)
	if err != nil {
		return err
	}
	daily := func(ds []trafficDay) [][]string {
		var rows [][]string
		for _, d := range ds {
			rows = append(rows, []string{day(d.Timestamp), itoa(d.Count), itoa(d.Uniques)})
		}
		return rows
	}
	if err := upsert(filepath.Join(dir, "views.csv"), viewCols, daily(views)); err != nil {
		return err
	}
	if err := upsert(filepath.Join(dir, "clones.csv"), cloneCols, daily(clones)); err != nil {
		return err
	}
	var refRows, pathRows [][]string
	for _, r := range refs {
		refRows = append(refRows, []string{today, r.Referrer, itoa(r.Count), itoa(r.Uniques)})
	}
	for _, p := range paths {
		pathRows = append(pathRows, []string{today, p.Path, p.Title, itoa(p.Count), itoa(p.Uniques)})
	}
	if err := upsert(filepath.Join(dir, "referrers.csv"), referrerCols, refRows); err != nil {
		return err
	}
	return upsert(filepath.Join(dir, "paths.csv"), popularPathCo, pathRows)
}

// downloadRows is one row per release asset. Kind separates what people
// install (archive) from what verifiers fetch: the GitHub Action downloads
// checksums.txt and its signature on every run, so "checksums" downloads
// roughly count Action runs plus manual verifications.
func downloadRows(today string, rels []release) [][]string {
	var rows [][]string
	for _, r := range rels {
		if r.Draft {
			continue
		}
		for _, a := range r.Assets {
			kind, platform := classify(a.Name)
			rows = append(rows, []string{today, r.Tag, strconv.FormatBool(r.Prerelease), kind, platform, itoa(a.Downloads)})
		}
	}
	return rows
}

// classify names a release asset's kind and, for archives and their SBOMs,
// its platform. Names follow .goreleaser.yaml:
// shorstop_<version>_<os>_<arch>.tar.gz (or .zip), plus .sbom.json.
func classify(name string) (kind, platform string) {
	switch name {
	case "checksums.txt":
		return "checksums", ""
	case "checksums.txt.sigstore.json":
		return "signature", ""
	}
	base, isSBOM := strings.CutSuffix(name, ".sbom.json")
	var ok bool
	for _, ext := range []string{".tar.gz", ".zip"} {
		if b, found := strings.CutSuffix(base, ext); found {
			base, ok = b, true
			break
		}
	}
	parts := strings.Split(base, "_")
	if !ok || len(parts) != 4 || parts[0] != "shorstop" {
		return "other", ""
	}
	platform = parts[2] + "_" + parts[3]
	if isSBOM {
		return "sbom", platform
	}
	return "archive", platform
}

// upsert replaces the rows of every date in rows, keeps all other dates, and
// writes the file sorted by date. Rows within a date are sorted too, so the
// file only changes when a number does.
func upsert(path string, header []string, rows [][]string) error {
	old, err := readCSV(path, header)
	if err != nil {
		return err
	}
	dates := map[string]bool{}
	for _, r := range rows {
		dates[r[0]] = true
	}
	all := slices.DeleteFunc(old, func(r []string) bool { return dates[r[0]] })
	all = append(all, rows...)
	slices.SortStableFunc(all, slices.Compare)

	f, err := os.CreateTemp(filepath.Dir(path), ".metrics-*.csv")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		f.Close()
		return err
	}
	if err := w.WriteAll(all); err != nil { // WriteAll flushes
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// readCSV returns a file's rows without its header. A missing file is empty.
// A header that doesn't match is an error: columns are never silently
// reinterpreted.
func readCSV(path string, header []string) ([][]string, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	recs, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(recs) == 0 {
		return nil, nil
	}
	if !slices.Equal(recs[0], header) {
		return nil, fmt.Errorf("%s: header is %v, want %v", path, recs[0], header)
	}
	return recs[1:], nil
}

// day is the date part of an API timestamp ("2026-11-16T00:00:00Z").
func day(timestamp string) string {
	d, _, _ := strings.Cut(timestamp, "T")
	return d
}

func itoa(n int) string { return strconv.Itoa(n) }
