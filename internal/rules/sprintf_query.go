package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(sprintfQuery{}) }

var queryMethods = map[string]bool{
	"Query": true, "QueryRow": true, "Exec": true,
	"QueryContext": true, "QueryRowContext": true, "ExecContext": true,
}

// sprintfQuery flags a database call whose statement is built with fmt.Sprintf.
// Formatting user input straight into SQL is the classic injection shape; the
// query should use placeholders and arguments instead.
type sprintfQuery struct{}

func (sprintfQuery) Name() string { return "sprintf-into-query" }

func (sprintfQuery) Describe() string {
	return "a SQL statement built with fmt.Sprintf and passed to Query/Exec — the shape of an injection"
}

func (s sprintfQuery) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !queryMethods[sel.Sel.Name] {
			return true
		}
		for _, arg := range call.Args {
			if isSprintf(arg) {
				file, line := pos(fn, call)
				out = append(out, Finding{
					Rule: s.Name(), Severity: "error", Func: fn.Name, File: file, Line: line,
					Message: sel.Sel.Name + " receives a fmt.Sprintf-built statement; use placeholders and arguments",
				})
				break
			}
		}
		return true
	})
	return out
}

func isSprintf(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "fmt" {
		return false
	}
	return sel.Sel.Name == "Sprintf" || sel.Sel.Name == "Sprint" || sel.Sel.Name == "Sprintln"
}
