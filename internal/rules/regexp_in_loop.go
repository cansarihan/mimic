package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(regexpInLoop{}) }

// regexpInLoop flags a regexp compiled inside a loop. Compilation is expensive
// and the pattern rarely changes per iteration, so it should be hoisted out.
type regexpInLoop struct{}

func (regexpInLoop) Name() string { return "regexp-in-loop" }

func (regexpInLoop) Describe() string {
	return "a regexp is compiled inside a loop; compile it once outside the loop instead"
}

func (r regexpInLoop) Match(fn srcwalk.Func) []Finding {
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
		for _, call := range directCalls(body) {
			if isRegexpCompile(call) {
				file, line := pos(fn, call)
				out = append(out, Finding{
					Rule: r.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
					Message: "regexp compiled inside a loop; hoist it above the loop",
				})
			}
		}
		return true
	})
	return out
}

func isRegexpCompile(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "regexp" {
		return false
	}
	return sel.Sel.Name == "MustCompile" || sel.Sel.Name == "Compile" || sel.Sel.Name == "CompilePOSIX"
}

// directCalls returns call expressions lexically inside body but not inside a
// nested loop or function literal.
func directCalls(body *ast.BlockStmt) []*ast.CallExpr {
	var found []*ast.CallExpr
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch t := n.(type) {
		case *ast.CallExpr:
			found = append(found, t)
		case *ast.ForStmt, *ast.RangeStmt, *ast.FuncLit:
			return
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
