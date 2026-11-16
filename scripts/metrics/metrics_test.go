package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	for name, want := range map[string][2]string{
		"checksums.txt":                                {"checksums", ""},
		"checksums.txt.sigstore.json":                  {"signature", ""},
		"shorstop_0.1.2_linux_amd64.tar.gz":            {"archive", "linux_amd64"},
		"shorstop_0.1.2_windows_amd64.zip":             {"archive", "windows_amd64"},
		"shorstop_0.1.0-rc.1_darwin_arm64.tar.gz":      {"archive", "darwin_arm64"},
		"shorstop_0.1.2_darwin_arm64.tar.gz.sbom.json": {"sbom", "darwin_arm64"},
		"notes.txt":                   {"other", ""},
		"shorstop_0.1.2_linux.tar.gz": {"other", ""},
	} {
		kind, platform := classify(name)
		if kind != want[0] || platform != want[1] {
			t.Errorf("classify(%q) = %q, %q; want %q, %q", name, kind, platform, want[0], want[1])
		}
	}
}

// TestUpsertReplacesADate checks that re-running a day replaces its rows,
// other days are kept, and an identical re-run leaves the file unchanged.
func TestUpsertReplacesADate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.csv")
	h := []string{"date", "n"}
	must(t, upsert(p, h, [][]string{{"2026-11-15", "1"}}))
	must(t, upsert(p, h, [][]string{{"2026-11-16", "2"}}))
	must(t, upsert(p, h, [][]string{{"2026-11-16", "3"}}))
	first := read(t, p)
	if want := "date,n\n2026-11-15,1\n2026-11-16,3\n"; first != want {
		t.Fatalf("got\n%s\nwant\n%s", first, want)
	}
	must(t, upsert(p, h, [][]string{{"2026-11-16", "3"}}))
	if again := read(t, p); again != first {
		t.Errorf("identical re-run changed the file:\n%s", again)
	}
}

func TestUpsertRejectsAnotherHeader(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.csv")
	must(t, os.WriteFile(p, []byte("date,other\n2026-11-15,1\n"), 0o644))
	if err := upsert(p, []string{"date", "n"}, nil); err == nil {
		t.Error("a file with a different header was accepted")
	}
}

// fakeAPI serves the endpoints the metrics read. Traffic needs the token.
func fakeAPI(t *testing.T, stars int) *httptest.Server {
	t.Helper()
	json200 := func(w http.ResponseWriter, v any) { _ = json.NewEncoder(w).Encode(v) }
	traffic := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer traffic" {
				http.Error(w, `{"message":"Must have push access"}`, http.StatusForbidden)
				return
			}
			h(w, r)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]int{"stargazers_count": stars, "forks_count": 1, "subscribers_count": 2, "open_issues_count": 0})
	})
	mux.HandleFunc("/repos/o/r/releases", func(w http.ResponseWriter, _ *http.Request) {
		json200(w, []release{
			{Tag: "v0.1.2", Assets: []asset{
				{Name: "checksums.txt", Downloads: 5},
				{Name: "shorstop_0.1.2_linux_amd64.tar.gz", Downloads: 4},
				{Name: "shorstop_0.1.2_darwin_arm64.tar.gz", Downloads: 2},
			}},
			{Tag: "v0.2.0", Draft: true, Assets: []asset{{Name: "checksums.txt", Downloads: 99}}},
		})
	})
	mux.HandleFunc("/repos/o/r/traffic/views", traffic(func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"views": []trafficDay{
			{Timestamp: "2026-11-15T00:00:00Z", Count: 10, Uniques: 3},
			{Timestamp: "2026-11-16T00:00:00Z", Count: 7, Uniques: 2},
		}})
	}))
	mux.HandleFunc("/repos/o/r/traffic/clones", traffic(func(w http.ResponseWriter, _ *http.Request) {
		json200(w, map[string]any{"clones": []trafficDay{{Timestamp: "2026-11-16T00:00:00Z", Count: 4, Uniques: 1}}})
	}))
	mux.HandleFunc("/repos/o/r/traffic/popular/referrers", traffic(func(w http.ResponseWriter, _ *http.Request) {
		json200(w, []referrer{{Referrer: "github.com", Count: 3, Uniques: 1}, {Referrer: "news.ycombinator.com", Count: 9, Uniques: 8}})
	}))
	mux.HandleFunc("/repos/o/r/traffic/popular/paths", traffic(func(w http.ResponseWriter, _ *http.Request) {
		json200(w, []popularPath{{Path: "/o/r", Title: "repo", Count: 12, Uniques: 9}})
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestRunWithoutTrafficToken(t *testing.T) {
	srv, dir := fakeAPI(t, 2), t.TempDir()
	c := &client{base: srv.URL, repo: "o/r"}
	must(t, run(context.Background(), c, dir, "2026-11-16"))

	if got, want := read(t, filepath.Join(dir, "repo.csv")), "date,stars,forks,watchers,open_issues\n2026-11-16,2,1,2,0\n"; got != want {
		t.Errorf("repo.csv:\n%s\nwant\n%s", got, want)
	}
	dl := read(t, filepath.Join(dir, "downloads.csv"))
	for _, want := range []string{
		"2026-11-16,v0.1.2,false,archive,linux_amd64,4",
		"2026-11-16,v0.1.2,false,checksums,,5",
	} {
		if !strings.Contains(dl, want) {
			t.Errorf("downloads.csv lacks %q:\n%s", want, dl)
		}
	}
	if strings.Contains(dl, "v0.2.0") {
		t.Error("a draft release was recorded")
	}
	if _, err := os.Stat(filepath.Join(dir, "views.csv")); err == nil {
		t.Error("views.csv written without a traffic token")
	}
	readme := read(t, filepath.Join(dir, "README.md"))
	for _, want := range []string{"| Stars | 2 | – | – |", "| Release archive downloads (all releases) | 6 |"} {
		if !strings.Contains(readme, want) {
			t.Errorf("README lacks %q:\n%s", want, readme)
		}
	}
}

func TestRunWithTrafficAndHistory(t *testing.T) {
	dir := t.TempDir()
	c := &client{base: fakeAPI(t, 2).URL, repo: "o/r", trafficToken: "traffic"}
	must(t, run(context.Background(), c, dir, "2026-11-09"))
	c.base = fakeAPI(t, 5).URL
	must(t, run(context.Background(), c, dir, "2026-11-16"))

	if got := read(t, filepath.Join(dir, "views.csv")); got != "date,views,unique_visitors\n2026-11-15,10,3\n2026-11-16,7,2\n" {
		t.Errorf("views.csv (re-fetched days must not duplicate):\n%s", got)
	}
	readme := read(t, filepath.Join(dir, "README.md"))
	for _, want := range []string{
		"| Stars | 5 | +3 | – |", // 7 days earlier there were 2
		"- Views: 17",
		"| news.ycombinator.com | 9 | 8 |", // most views first
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README lacks %q:\n%s", want, readme)
		}
	}
	if strings.Index(readme, "news.ycombinator.com") > strings.Index(readme, "| github.com |") {
		t.Error("referrers aren't sorted by views")
	}
}

func TestRunFailsOnBadTrafficToken(t *testing.T) {
	c := &client{base: fakeAPI(t, 2).URL, repo: "o/r", trafficToken: "expired"}
	err := run(context.Background(), c, t.TempDir(), "2026-11-16")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want the 403 reported, so an expired token is noticed", err)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	must(t, err)
	return string(b)
}
