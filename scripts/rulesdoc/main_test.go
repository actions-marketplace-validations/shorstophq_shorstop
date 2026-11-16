package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shorstophq/shorstop/internal/detect/cert"
	"github.com/shorstophq/shorstop/internal/detect/config"
	"github.com/shorstophq/shorstop/internal/rules"
)

// TestEveryRuleListed checks that the reference lists every rule exactly once
// per place it applies, and that every table row has the right cell count
// (an unescaped | in a description or example would add a cell).
func TestEveryRuleListed(t *testing.T) {
	pack, err := rules.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	write(&b, pack, cert.RuleDescriptions, config.RuleDescriptions)
	out := b.String()

	ids := map[string]bool{}
	for _, r := range pack.Rules {
		ids[r.ID] = true
	}
	for id := range cert.RuleDescriptions {
		ids[id] = true
	}
	for id := range config.RuleDescriptions {
		ids[id] = true
	}
	for id := range ids {
		if !strings.Contains(out, "| `"+id+"` |") {
			t.Errorf("rule %s is missing", id)
		}
	}

	var cols int
	for line := range strings.Lines(out) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			cols = 0
			continue
		}
		n := strings.Count(strings.ReplaceAll(line, `\|`, ""), "|")
		if cols == 0 {
			cols = n
		} else if n != cols {
			t.Errorf("row has %d separators, header has %d: %s", n, cols, line)
		}
	}
}

func TestCode(t *testing.T) {
	for in, want := range map[string]string{
		"a|b":     "`a\\|b`",
		"x `y` z": "``x `y` z``",
		"`edge`":  "`` `edge` ``",
		"plain":   "`plain`",
		"":        "",
	} {
		if got := code(in); got != want {
			t.Errorf("code(%q) = %q, want %q", in, got, want)
		}
	}
}
