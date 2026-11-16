package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// summarize renders README.md for the metrics branch: the latest numbers with
// their change over 7 and 30 days, computed from the CSV files alone.
func summarize(dir, repo string) (string, error) {
	repoRows, err := readCSV(filepath.Join(dir, "repo.csv"), repoCols)
	if err != nil {
		return "", err
	}
	dl, err := readCSV(filepath.Join(dir, "downloads.csv"), downloadCols)
	if err != nil {
		return "", err
	}
	views, err := readCSV(filepath.Join(dir, "views.csv"), viewCols)
	if err != nil {
		return "", err
	}
	clones, err := readCSV(filepath.Join(dir, "clones.csv"), cloneCols)
	if err != nil {
		return "", err
	}
	refs, err := readCSV(filepath.Join(dir, "referrers.csv"), referrerCols)
	if err != nil {
		return "", err
	}
	if len(repoRows) == 0 {
		return "", fmt.Errorf("repo.csv is empty")
	}
	latest := repoRows[len(repoRows)-1][0]

	// series maps date -> value for one metric.
	series := map[string]map[string]int{
		"Stars":    column(repoRows, 1),
		"Forks":    column(repoRows, 2),
		"Watchers": column(repoRows, 3),
		"Release archive downloads (all releases)":                sumBy(dl, func(r []string) bool { return r[3] == "archive" }),
		"checksums.txt downloads (≈ Action runs + verifications)": sumBy(dl, func(r []string) bool { return r[3] == "checksums" }),
	}
	order := []string{"Stars", "Forks", "Watchers",
		"Release archive downloads (all releases)",
		"checksums.txt downloads (≈ Action runs + verifications)"}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s metrics\n\n", repo)
	fmt.Fprintf(&b, "Recorded daily by `.github/workflows/metrics.yml` on `main` (`scripts/metrics`). Latest: **%s** (UTC).\n\n", latest)
	b.WriteString("| Metric | Now | 7 days | 30 days |\n|---|---:|---:|---:|\n")
	for _, name := range order {
		s := series[name]
		now, ok := s[latest]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "| %s | %d | %s | %s |\n", name, now, delta(s, latest, 7), delta(s, latest, 30))
	}

	if len(views) > 0 {
		b.WriteString("\n## Traffic, last 14 days\n\n")
		// Daily unique counts can't be summed (a returning visitor would be
		// counted once per day), so only totals are shown; views.csv and
		// clones.csv have each day's uniques.
		fmt.Fprintf(&b, "- Views: %d\n", sumLast(views, latest, 14, 1))
		fmt.Fprintf(&b, "- Clones: %d\n", sumLast(clones, latest, 14, 1))
	}
	var top []([]string)
	for _, r := range refs {
		if r[0] == latest {
			top = append(top, r)
		}
	}
	if len(top) > 0 {
		slices.SortStableFunc(top, func(a, b []string) int { return atoi(b[2]) - atoi(a[2]) })
		b.WriteString("\n## Top referrers, last 14 days\n\n| Referrer | Views | Unique visitors |\n|---|---:|---:|\n")
		for _, r := range top {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", r[1], r[2], r[3])
		}
	}
	b.WriteString("\nNot measured: Docker image pulls (ghcr.io publishes no pull count) and Action installs (the Marketplace publishes none). " +
		"Homebrew installs are inside the archive downloads, since the cask downloads the release archive.\n")
	return b.String(), nil
}

func column(rows [][]string, i int) map[string]int {
	m := map[string]int{}
	for _, r := range rows {
		m[r[0]] = atoi(r[i])
	}
	return m
}

// sumBy totals the downloads column per date over the rows keep accepts.
func sumBy(rows [][]string, keep func([]string) bool) map[string]int {
	m := map[string]int{}
	for _, r := range rows {
		if keep(r) {
			m[r[0]] += atoi(r[5])
		}
	}
	return m
}

// delta is the change since n days before latest, or "–" when there's no
// record that old.
func delta(s map[string]int, latest string, n int) string {
	t, err := time.Parse(time.DateOnly, latest)
	if err != nil {
		return "–"
	}
	then, ok := s[t.AddDate(0, 0, -n).Format(time.DateOnly)]
	if !ok {
		return "–"
	}
	return fmt.Sprintf("%+d", s[latest]-then)
}

// sumLast totals column col over the n days ending at latest.
func sumLast(rows [][]string, latest string, n, col int) int {
	t, err := time.Parse(time.DateOnly, latest)
	if err != nil {
		return 0
	}
	from := t.AddDate(0, 0, -(n - 1)).Format(time.DateOnly)
	total := 0
	for _, r := range rows {
		if r[0] >= from && r[0] <= latest {
			total += atoi(r[col])
		}
	}
	return total
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
