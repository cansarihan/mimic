package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(hardcodedCreds{}) }

// hardcodedCreds flags a secret-named variable assigned a non-empty string
// literal — a password, token or key baked into the source.
type hardcodedCreds struct{}

func (hardcodedCreds) Name() string { return "hardcoded-credentials" }

func (hardcodedCreds) Describe() string {
	return "a secret-named variable is assigned a string literal; keep credentials out of source"
}

func (h hardcodedCreds) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i := 0; i < len(assign.Lhs) && i < len(assign.Rhs); i++ {
			ident, ok := assign.Lhs[i].(*ast.Ident)
			if !ok || !secretName(ident.Name) {
				continue
			}
			if lit, ok := assign.Rhs[i].(*ast.BasicLit); ok && lit.Kind == token.STRING && len(lit.Value) > 2 {
				file, line := pos(fn, assign)
				out = append(out, Finding{
					Rule: h.Name(), Severity: "error", Func: fn.Name, File: file, Line: line,
					Message: ident.Name + " is assigned a hardcoded string literal",
				})
			}
		}
		return true
	})
	return out
}

func secretName(name string) bool {
	lower := strings.ToLower(name)
	for _, w := range secretWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}
