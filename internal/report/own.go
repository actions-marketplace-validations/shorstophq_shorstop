package report

import (
	"bytes"
	"regexp"
)

// ToolName is how ShorStop names itself in SARIF and CBOM output.
const ToolName = "shorstop"

// markdownHeading starts every Markdown report.
const markdownHeading = "## ShorStop post-quantum scan"

// Patterns for the start of ShorStop's own SARIF and CBOM output, with any
// whitespace between tokens, so a report reformatted by jq or minified still
// matches. They're anchored on the tool's own name in the place each format
// records the producing tool, so SARIF or CycloneDX files from other tools
// are still scanned.
var (
	ownSARIF = regexp.MustCompile(`"driver"\s*:\s*\{\s*"name"\s*:\s*"` + ToolName + `"`)
	ownCBOM  = regexp.MustCompile(`"tools"\s*:\s*\{\s*"components"\s*:\s*\[\s*\{\s*"type"\s*:\s*"application"\s*,\s*"name"\s*:\s*"` + ToolName + `"`)
)

// IsOwnReport reports whether head, the start of a file, is a report that
// ShorStop wrote: SARIF, a CBOM (--format cbom or json) or Markdown. Such
// a file names every algorithm it found, and scanning it would report them
// all again. The identifying part is within the first few hundred bytes of
// each format, so head can be short.
func IsOwnReport(head []byte) bool {
	head = bytes.TrimPrefix(head, []byte("\xef\xbb\xbf")) // a UTF-8 BOM some editors add
	trimmed := bytes.TrimLeft(head, " \t\r\n")
	if bytes.HasPrefix(trimmed, []byte(markdownHeading+"\n")) || bytes.HasPrefix(trimmed, []byte(markdownHeading+"\r\n")) {
		return true
	}
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	if bytes.Contains(head, []byte(`"bomFormat"`)) {
		return ownCBOM.Match(head)
	}
	return bytes.Contains(head, []byte("sarif")) && ownSARIF.Match(head)
}
