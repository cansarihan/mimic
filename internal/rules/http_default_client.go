package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(httpDefaultClient{}) }

var defaultClientCalls = map[string]bool{"Get": true, "Post": true, "Head": true, "PostForm": true}

// httpDefaultClient flags a request made through net/http's package-level
// helpers. They use http.DefaultClient, which has no timeout, so a slow or
// hostile server can hang the caller forever.
type httpDefaultClient struct{}

func (httpDefaultClient) Name() string { return "http-default-client" }

func (httpDefaultClient) Describe() string {
	return "http.Get/Post use the default client, which has no timeout; use a client with a Timeout"
}

func (h httpDefaultClient) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !defaultClientCalls[sel.Sel.Name] {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "http" {
			file, line := pos(fn, call)
			out = append(out, Finding{
				Rule: h.Name(), Severity: "warning", Func: fn.Name, File: file, Line: line,
				Message: "http." + sel.Sel.Name + " uses a client with no timeout",
			})
		}
		return true
	})
	return out
}
