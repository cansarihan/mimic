package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(osExitInLibrary{}) }

// osExitInLibrary flags os.Exit outside package main. A library that exits the
// process skips every deferred cleanup and takes the decision away from its
// caller; it should return an error instead.
type osExitInLibrary struct{}

func (osExitInLibrary) Name() string { return "os-exit-in-library" }

func (osExitInLibrary) Describe() string {
	return "os.Exit outside package main skips deferred cleanup and robs the caller of the decision; return an error"
}

func (o osExitInLibrary) Match(fn srcwalk.Func) []Finding {
	if fn.Package == "main" {
		return nil
	}
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Exit" {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "os" {
			file, line := pos(fn, call)
			out = append(out, Finding{
				Rule: o.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
				Message: "os.Exit in package " + fn.Package + "; return an error to the caller",
			})
		}
		return true
	})
	return out
}
