package rules

import (
	"go/ast"
	"strings"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(randSecret{}) }

var secretWords = []string{"token", "secret", "key", "password", "passwd", "nonce", "salt", "session", "csrf", "otp", "apikey"}

// randSecret flags a function that looks like it mints a secret yet draws from
// math/rand, which is predictable. Such code should use crypto/rand.
type randSecret struct{}

func (randSecret) Name() string { return "math-rand-secret" }

func (randSecret) Describe() string {
	return "a secret-sounding function draws from math/rand; use crypto/rand for anything security-sensitive"
}

func (rs randSecret) Match(fn srcwalk.Func) []Finding {
	lower := strings.ToLower(fn.Name)
	secretish := false
	for _, w := range secretWords {
		if strings.Contains(lower, w) {
			secretish = true
			break
		}
	}
	if !secretish {
		return nil
	}

	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "rand" && mathRandFunc(sel.Sel.Name) {
			file, line := pos(fn, call)
			out = append(out, Finding{
				Rule: rs.Name(), Severity: "error", Func: fn.Name, File: file, Line: line,
				Message: fn.Name + " builds a secret with math/rand's " + sel.Sel.Name + "; use crypto/rand",
			})
			return false
		}
		return true
	})
	return out
}

func mathRandFunc(name string) bool {
	switch name {
	case "Int", "Intn", "Int31", "Int31n", "Int63", "Int63n", "Uint32", "Uint64", "Read", "Float64", "Perm", "Shuffle":
		return true
	default:
		return false
	}
}
