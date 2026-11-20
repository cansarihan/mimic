// Package rules holds the built-in bug patterns Mimic looks for. Each rule is a
// matcher that inspects a function's AST and reports the shapes that are wrong
// often enough to be worth a second look — a defer inside a loop, a discarded
// append, a query built with Sprintf. The rules need no type information, so
// they run on any Go source that parses.
package rules

import (
	"go/ast"
	"sort"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

// Finding is one rule hit at a position in the source.
type Finding struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Func     string `json:"func"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

// Matcher inspects one function and reports findings.
type Matcher interface {
	Name() string
	Describe() string
	Match(fn srcwalk.Func) []Finding
}

var registry []Matcher

func register(m Matcher) { registry = append(registry, m) }

// Registry returns every built-in matcher, sorted by name so the catalogue and
// the SARIF rule list are stable from run to run.
func Registry() []Matcher {
	out := make([]Matcher, len(registry))
	copy(out, registry)
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// pos is a small helper to turn an AST node into a file and line.
func pos(fn srcwalk.Func, node ast.Node) (string, int) {
	p := fn.Fset.Position(node.Pos())
	return p.Filename, p.Line
}
