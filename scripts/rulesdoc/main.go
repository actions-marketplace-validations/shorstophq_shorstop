// Command rulesdoc writes the docs site's rules reference (Markdown) from the
// rules ShorStop ships: the built-in rule pack and the certificate and
// configuration detectors' rules. The page is generated at build time and
// never committed, so it can't drift from the code.
//
//	go run ./scripts/rulesdoc > site/content/docs/rules.md
package main

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/shorstophq/shorstop/internal/detect/cert"
	"github.com/shorstophq/shorstop/internal/detect/config"
	"github.com/shorstophq/shorstop/internal/rules"
)

func main() {
	pack, err := rules.Builtin()
	if err != nil {
		fmt.Fprintln(os.Stderr, "rulesdoc:", err)
		os.Exit(1)
	}
	w := bufio.NewWriter(os.Stdout)
	write(w, pack, cert.RuleDescriptions, config.RuleDescriptions)
	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "rulesdoc:", err)
		os.Exit(1)
	}
}

// languageNames are the headings for rules.Languages keys.
var languageNames = map[string]string{
	"c": "C", "cpp": "C++", "csharp": "C#", "go": "Go", "java": "Java",
	"javascript": "JavaScript", "kotlin": "Kotlin", "php": "PHP",
	"python": "Python", "ruby": "Ruby", "rust": "Rust", "typescript": "TypeScript",
}

// ecosystemNames are the headings for rules.Ecosystems.
var ecosystemNames = map[string]string{
	"cargo": "Cargo (Rust)", "composer": "Composer (PHP)", "gem": "RubyGems",
	"golang": "Go modules", "maven": "Maven and Gradle", "npm": "npm",
	"nuget": "NuGet", "pypi": "PyPI",
}

func write(w io.Writer, pack *rules.Pack, certRules, configRules map[string]string) {
	byDetector := map[string][]*rules.Rule{}
	for i := range pack.Rules {
		r := &pack.Rules[i]
		byDetector[r.Detector] = append(byDetector[r.Detector], r)
	}
	total := len(pack.Rules) + len(certRules) + len(configRules)
	regexLangs := map[string]bool{}
	for _, r := range byDetector["regex"] {
		for _, l := range r.Languages {
			regexLangs[l] = true
		}
	}
	ecosystems := map[string]bool{}
	for _, r := range byDetector["deps"] {
		ecosystems[r.Match.Ecosystem] = true
	}

	fmt.Fprintf(w, `---
title: Rules reference
weight: 60
editURL: https://github.com/shorstophq/shorstop/edit/main/internal/rules/rules.yaml
---

ShorStop ships **%d rules**. This page is generated from the rules in the
release, so it always matches what %s prints.

A rule's ID appears in every finding, in `+"`shorstop:ignore`"+` comments and in
`+"`disable:`"+` in `+"`.shorstop.yaml`"+`. IDs never change meaning: a rule that has to
change is retired and replaced by a new ID. Severity isn't part of a rule: it
comes from what was found (see [How findings are rated](../ratings/)).

| Detector | Rules | Looks at |
|---|---:|---|
| Go source (`+"`goast`"+`) | %d | `+"`.go`"+` files, parsed |
| Patterns (`+"`regex`"+`) | %d | source code in %d languages, and any text file |
| Dependencies (`+"`deps`"+`) | %d | manifests and lockfiles in %d ecosystems |
| Certificates and keys (`+"`cert`"+`) | %d | PEM, DER, PKCS#12 and OpenSSH files |
| Server configuration (`+"`config`"+`) | %d | TLS, SSH and OpenSSL settings, Kubernetes Secrets |
`, total, "`shorstop rules list`",
		len(byDetector["goast"]), len(byDetector["regex"]), len(regexLangs),
		len(byDetector["deps"]), len(ecosystems), len(certRules), len(configRules))

	// Go AST.
	fmt.Fprint(w, "\n## Go source\n\nMatched on the parsed syntax tree, so comments and strings never match. Key sizes are read from the call when they're integer literals.\n\n")
	fmt.Fprint(w, "| Rule | Algorithm | Matches | Description |\n|---|---|---|---|\n")
	for _, r := range sorted(byDetector["goast"]) {
		m := r.Match.Import
		if r.Match.Call != "" {
			m += "." + r.Match.Call + "()"
		}
		fmt.Fprintf(w, "| `%s` | %s | %s | %s |\n", r.ID, alg(r), code(m), cell(r.Description))
	}

	// Regex, one section per language; rules without languages apply to every text file.
	fmt.Fprint(w, "\n## Patterns\n\nMatched line by line. Every rule lists example lines that it must match, and those examples are checked each time the rules load: they're shown here as what the rule finds.\n")
	groups := map[string][]*rules.Rule{}
	for _, r := range byDetector["regex"] {
		if len(r.Languages) == 0 {
			groups[""] = append(groups[""], r)
		}
		for _, l := range r.Languages {
			groups[l] = append(groups[l], r)
		}
	}
	langs := slices.SortedFunc(maps.Keys(groups), func(a, b string) int {
		return cmp.Compare(langTitle(a), langTitle(b))
	})
	// Any-file rules last: they're the catch-all.
	if i := slices.Index(langs, ""); i >= 0 {
		langs = append(slices.Delete(langs, i, i+1), "")
	}
	for _, l := range langs {
		fmt.Fprintf(w, "\n### %s\n\n", langTitle(l))
		if l != "" {
			fmt.Fprintf(w, "Files: %s.\n\n", exts(rules.Languages[l]))
		}
		fmt.Fprint(w, "| Rule | Algorithm | Example | Description |\n|---|---|---|---|\n")
		for _, r := range sorted(groups[l]) {
			ex := ""
			if len(r.Match.Examples) > 0 {
				ex = code(r.Match.Examples[0])
			}
			fmt.Fprintf(w, "| `%s` | %s | %s | %s |\n", r.ID, alg(r), ex, cell(r.Description))
		}
	}

	// Dependencies, one section per ecosystem.
	fmt.Fprint(w, "\n## Dependencies\n\nA dependency is reported where it's declared: the manifest or lockfile line. A trailing `*` after a separator in a package name matches any suffix.\n")
	eco := map[string][]*rules.Rule{}
	for _, r := range byDetector["deps"] {
		eco[r.Match.Ecosystem] = append(eco[r.Match.Ecosystem], r)
	}
	for _, e := range slices.SortedFunc(maps.Keys(eco), func(a, b string) int {
		return cmp.Compare(strings.ToLower(ecosystemNames[a]), strings.ToLower(ecosystemNames[b]))
	}) {
		fmt.Fprintf(w, "\n### %s\n\n| Rule | Package | Algorithm | Description |\n|---|---|---|---|\n", cmp.Or(ecosystemNames[e], e))
		for _, r := range sorted(eco[e]) {
			fmt.Fprintf(w, "| `%s` | %s | %s | %s |\n", r.ID, code(r.Match.Package), alg(r), cell(r.Description))
		}
	}

	// Code-defined rules: the algorithm comes from what's parsed, not the rule.
	fmt.Fprint(w, "\n## Certificates and keys\n\nThe algorithm, key size and signature hash are read from each certificate or key, so one rule covers them all.\n\n")
	described(w, certRules)
	fmt.Fprint(w, "\n## Server configuration\n\nOne rule per kind of setting. The algorithms come from the setting's value, with removals such as `!3DES` applied first, so only what's actually enabled is reported.\n\n")
	described(w, configRules)
}

func described(w io.Writer, m map[string]string) {
	fmt.Fprint(w, "| Rule | Description |\n|---|---|\n")
	for _, id := range slices.Sorted(maps.Keys(m)) {
		fmt.Fprintf(w, "| `%s` | %s |\n", id, cell(m[id]))
	}
}

func sorted(rs []*rules.Rule) []*rules.Rule {
	out := slices.Clone(rs)
	slices.SortFunc(out, func(a, b *rules.Rule) int { return cmp.Compare(a.ID, b.ID) })
	return out
}

func langTitle(l string) string {
	if l == "" {
		return "Any text file"
	}
	return cmp.Or(languageNames[l], l)
}

func exts(es []string) string {
	q := make([]string, len(es))
	for i, e := range es {
		q[i] = "`" + e + "`"
	}
	return strings.Join(q, ", ")
}

// alg is the rule's algorithm, with key size and mode when the rule fixes them.
func alg(r *rules.Rule) string {
	a := cmp.Or(r.Algorithm, "–")
	// X25519 and Ed25519 name their curve; "-255" would only confuse.
	if r.KeySize > 0 && r.Algorithm != "X25519" && r.Algorithm != "EdDSA" {
		a += fmt.Sprintf("-%d", r.KeySize)
	}
	if r.Mode != "" {
		a += "-" + r.Mode
	}
	return cell(a)
}

// code renders s as inline code in a table cell. Backticks inside s would end
// the span, so a longer fence is used; pipes would end the cell, so they're
// escaped (GFM tables handle \| before inline parsing, even in code).
func code(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "|", `\|`)
	fence := "`"
	for strings.Contains(s, fence) {
		fence += "`"
	}
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		s = " " + s + " "
	}
	return fence + s + fence
}

// cell makes plain text safe in a table cell.
func cell(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.ReplaceAll(s, "\n", " ")
}
