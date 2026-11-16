package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestIsOwnReportRecognisesEveryFormat renders each report from the golden
// findings and checks it's recognised from its first 8000 bytes (what the
// crawler sees), also after jq-style reformatting.
func TestIsOwnReportRecognisesEveryFormat(t *testing.T) {
	fs := goldenFindings()
	var sarif, cbomOut, md bytes.Buffer
	if err := WriteSARIF(&sarif, fs, SARIFOptions{ToolVersion: "v0.0.0-test"}); err != nil {
		t.Fatal(err)
	}
	if err := WriteCBOM(&cbomOut, CBOM(fs, CBOMOptions{ToolVersion: "v0.0.0-test"})); err != nil {
		t.Fatal(err)
	}
	if err := WriteMarkdown(&md, fs, MarkdownOptions{}); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{
		"sarif":            sarif.Bytes(),
		"cbom":             cbomOut.Bytes(),
		"markdown":         md.Bytes(),
		"sarif, compacted": compact(t, sarif.Bytes()),
		"cbom, compacted":  compact(t, cbomOut.Bytes()),
		"sarif, with BOM":  append([]byte("\xef\xbb\xbf"), sarif.Bytes()...),
		"markdown, CRLF":   bytes.ReplaceAll(md.Bytes(), []byte("\n"), []byte("\r\n")),
	}
	for name, out := range cases {
		if !IsOwnReport(out[:min(len(out), 8000)]) {
			t.Errorf("%s: not recognised; starts:\n%.300s", name, out)
		}
	}
}

// TestIsOwnReportLeavesOtherFilesAlone: other tools' SARIF and CBOMs, and
// files that merely mention ShorStop, are still scanned.
func TestIsOwnReportLeavesOtherFilesAlone(t *testing.T) {
	for name, s := range map[string]string{
		"another tool's SARIF":                 `{"$schema":"https://json.schemastore.org/sarif-2.1.0.json","version":"2.1.0","runs":[{"tool":{"driver":{"name":"semgrep"}}}]}`,
		"another tool's CBOM":                  `{"bomFormat":"CycloneDX","specVersion":"1.6","metadata":{"tools":{"components":[{"type":"application","name":"cbomkit"}]}}}`,
		"a CBOM listing shorstop as a library": `{"bomFormat":"CycloneDX","components":[{"type":"library","name":"shorstop"}]}`,
		"JSON config naming shorstop":          `{"driver":{"name":"shorstop"}}`,
		"Markdown quoting the heading":         "# Notes\n\n## ShorStop post-quantum scan\n",
		"Markdown with a longer heading":       "## ShorStop post-quantum scan results\n",
		"Go source":                            "package main\n// \"driver\": {\"name\": \"shorstop\"} sarif\n",
		"empty":                                "",
	} {
		if IsOwnReport([]byte(s)) {
			t.Errorf("%s: wrongly taken for a ShorStop report", name)
		}
	}
}

func compact(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, b); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "\n") {
		t.Fatal("not compacted")
	}
	return out.Bytes()
}
