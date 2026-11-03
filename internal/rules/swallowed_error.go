package rules

import (
	"go/ast"
	"go/token"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(swallowedError{}) }

// swallowedError flags an `if err != nil { }` whose body is empty: the error was
// checked and then thrown away, which hides the failure.
type swallowedError struct{}

func (swallowedError) Name() string { return "swallowed-error" }

func (swallowedError) Describe() string {
	return "an error is checked with `if err != nil` but the block is empty, so the failure is silently dropped"
}

func (s swallowedError) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok || len(ifStmt.Body.List) != 0 {
			return true
		}
		if isErrNotNil(ifStmt.Cond) {
			file, line := pos(fn, ifStmt)
			out = append(out, Finding{
				Rule: s.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
				Message: "empty `if err != nil` block swallows the error",
			})
		}
		return true
	})
	return out
}

func isErrNotNil(cond ast.Expr) bool {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	left, ok := bin.X.(*ast.Ident)
	if !ok || left.Name != "err" {
		return false
	}
	right, ok := bin.Y.(*ast.Ident)
	return ok && right.Name == "nil"
}
