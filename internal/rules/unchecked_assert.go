package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(uncheckedAssert{}) }

// uncheckedAssert flags a single-value type assertion x.(T). If the dynamic type
// is not T the program panics; the comma-ok form v, ok := x.(T) is safe.
type uncheckedAssert struct{}

func (uncheckedAssert) Name() string { return "unchecked-type-assert" }

func (uncheckedAssert) Describe() string {
	return "a single-value type assertion x.(T) panics on a type mismatch; use the comma-ok form"
}

func (u uncheckedAssert) Match(fn srcwalk.Func) []Finding {
	// First collect the assertions that appear as the sole RHS of a two-value
	// assignment — those are the safe comma-ok form.
	safe := map[ast.Node]struct{}{}
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if ok && len(assign.Lhs) == 2 && len(assign.Rhs) == 1 {
			if ta, ok := assign.Rhs[0].(*ast.TypeAssertExpr); ok {
				safe[ta] = struct{}{}
			}
		}
		return true
	})

	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		ta, ok := n.(*ast.TypeAssertExpr)
		if !ok || ta.Type == nil { // Type == nil is the x.(type) switch guard
			return true
		}
		if _, isSafe := safe[ta]; isSafe {
			return true
		}
		file, line := pos(fn, ta)
		out = append(out, Finding{
			Rule: u.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
			Message: "unchecked type assertion panics on a mismatch; use v, ok := x.(T)",
		})
		return true
	})
	return out
}
