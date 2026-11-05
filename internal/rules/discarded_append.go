package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(discardedAppend{}) }

// discardedAppend flags an append whose result is thrown away — assigned to the
// blank identifier. append may reallocate, so discarding its result loses the
// appended element. A bare `append(s, x)` statement does not compile, so the
// mistake that survives a build is `_ = append(s, x)`.
type discardedAppend struct{}

func (discardedAppend) Name() string { return "discarded-append" }

func (discardedAppend) Describe() string {
	return "the result of append is assigned to _; append can reallocate, so the appended element is lost"
}

func (d discardedAppend) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || !allBlank(assign.Lhs) {
			return true
		}
		for _, rhs := range assign.Rhs {
			if isAppendCall(rhs) {
				file, line := pos(fn, assign)
				out = append(out, Finding{
					Rule: d.Name(), Severity: "error", Func: fn.Name, File: file, Line: line,
					Message: "append result is discarded; assign it back to the slice",
				})
				break
			}
		}
		return true
	})
	return out
}

func allBlank(exprs []ast.Expr) bool {
	for _, e := range exprs {
		ident, ok := e.(*ast.Ident)
		if !ok || ident.Name != "_" {
			return false
		}
	}
	return len(exprs) > 0
}

func isAppendCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "append"
}
