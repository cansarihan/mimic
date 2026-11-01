package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(discardedAppend{}) }

// discardedAppend flags an append whose result is thrown away. append may return
// a new slice header, so `append(xs, v)` as a statement loses the appended data.
type discardedAppend struct{}

func (discardedAppend) Name() string { return "discarded-append" }

func (discardedAppend) Describe() string {
	return "the result of append is discarded; append can reallocate, so the appended element is lost"
}

func (d discardedAppend) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "append" {
			file, line := pos(fn, call)
			out = append(out, Finding{
				Rule: d.Name(), Severity: "error", Func: fn.Name, File: file, Line: line,
				Message: "append result is discarded; assign it back to the slice",
			})
		}
		return true
	})
	return out
}
