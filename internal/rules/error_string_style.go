package rules

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(errorStringStyle{}) }

// errorStringStyle flags an error message that starts with a capital letter or
// ends with punctuation. Go errors are wrapped and joined into longer messages,
// so by convention they are lower-case fragments without a trailing period.
type errorStringStyle struct{}

func (errorStringStyle) Name() string { return "error-string-style" }

func (errorStringStyle) Describe() string {
	return "an error string starts with a capital or ends with punctuation; errors are wrapped, so keep them lower-case fragments"
}

func (e errorStringStyle) Match(fn srcwalk.Func) []Finding {
	var out []Finding
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if !(pkg.Name == "errors" && sel.Sel.Name == "New") && !(pkg.Name == "fmt" && sel.Sel.Name == "Errorf") {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		msg, err := strconv.Unquote(lit.Value)
		if err != nil || msg == "" {
			return true
		}
		first := []rune(msg)[0]
		last := msg[len(msg)-1]
		if unicode.IsUpper(first) && !startsWithAcronym(msg) || strings.ContainsRune(".!:\n", rune(last)) {
			file, line := pos(fn, call)
			out = append(out, Finding{
				Rule: e.Name(), Severity: "info", Func: fn.Name, File: file, Line: line,
				Message: "error string " + strconv.Quote(msg) + " should be a lower-case fragment without trailing punctuation",
			})
		}
		return true
	})
	return out
}

// startsWithAcronym allows messages that open with an all-caps word such as
// "HTTP" or "EOF", which are legitimately capitalized.
func startsWithAcronym(msg string) bool {
	word := msg
	if i := strings.IndexAny(msg, " :"); i > 0 {
		word = msg[:i]
	}
	if len(word) < 2 {
		return false
	}
	for _, r := range word {
		if unicode.IsLower(r) {
			return false
		}
	}
	return true
}
