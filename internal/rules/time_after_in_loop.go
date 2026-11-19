package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(timeAfterInLoop{}) }

// timeAfterInLoop flags time.After inside a loop. Each call allocates a timer
// that is not released until it fires, so a select loop that rarely times out
// piles up timers; reuse a time.Timer instead.
type timeAfterInLoop struct{}

func (timeAfterInLoop) Name() string { return "time-after-in-loop" }

func (timeAfterInLoop) Describe() string {
	return "time.After inside a loop allocates a timer per iteration that lives until it fires; reuse a time.Timer"
}

func (t timeAfterInLoop) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		var body *ast.BlockStmt
		switch loop := n.(type) {
		case *ast.ForStmt:
			body = loop.Body
		case *ast.RangeStmt:
			body = loop.Body
		default:
			return true
		}
		for _, call := range directCalls(body) {
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "After" {
				continue
			}
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "time" {
				file, line := pos(fn, call)
				out = append(out, Finding{
					Rule: t.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
					Message: "time.After in a loop leaks a timer each iteration until it fires",
				})
			}
		}
		return true
	})
	return out
}
