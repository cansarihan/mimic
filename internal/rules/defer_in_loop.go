package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(deferInLoop{}) }

// deferInLoop flags a defer whose nearest enclosing control structure is a loop.
// Deferred calls run when the function returns, not when the iteration ends, so
// a defer in a loop holds every resource until the whole function finishes.
type deferInLoop struct{}

func (deferInLoop) Name() string { return "defer-in-loop" }

func (deferInLoop) Describe() string {
	return "a defer inside a loop runs only when the function returns, so it leaks the resource on every iteration"
}

func (d deferInLoop) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		var body *ast.BlockStmt
		switch t := n.(type) {
		case *ast.ForStmt:
			body = t.Body
		case *ast.RangeStmt:
			body = t.Body
		default:
			return true
		}
		for _, stmt := range directDefers(body) {
			file, line := pos(fn, stmt)
			out = append(out, Finding{
				Rule: d.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
				Message: "defer inside a loop; the resource is released only when " + fn.Name + " returns",
			})
		}
		return true
	})
	return out
}

// directDefers returns the defer statements lexically inside body but not inside
// a nested loop or function literal, which have their own scope.
func directDefers(body *ast.BlockStmt) []*ast.DeferStmt {
	var found []*ast.DeferStmt
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch t := n.(type) {
		case *ast.DeferStmt:
			found = append(found, t)
		case *ast.ForStmt, *ast.RangeStmt, *ast.FuncLit:
			return // a nested scope handles its own defers
		}
		for _, child := range children(n) {
			walk(child)
		}
	}
	for _, stmt := range body.List {
		walk(stmt)
	}
	return found
}

func children(n ast.Node) []ast.Node {
	var kids []ast.Node
	ast.Inspect(n, func(c ast.Node) bool {
		if c == nil || c == n {
			return c == n
		}
		kids = append(kids, c)
		return false
	})
	return kids
}
