// Package fingerprint reduces a function to a structural token stream that
// ignores names and literal values, so two functions that share a shape — a
// copy-paste and its edited sibling — look alike even when every identifier was
// renamed. It is the basis for finding the variants of a known-buggy function.
package fingerprint

import (
	"go/ast"
	"go/token"
)

// Tokens walks a function body and emits one token per AST node, keyed on the
// node's kind and, for the nodes where structure depends on it, its operator.
// Identifiers and literals collapse to a single placeholder token so renaming a
// variable or changing a constant does not change the fingerprint.
func Tokens(decl *ast.FuncDecl) []string {
	if decl == nil || decl.Body == nil {
		return nil
	}
	var tokens []string
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if tok := tokenFor(n); tok != "" {
			tokens = append(tokens, tok)
		}
		return true
	})
	return tokens
}

func tokenFor(n ast.Node) string {
	switch node := n.(type) {
	case *ast.Ident:
		return "id"
	case *ast.BasicLit:
		return "lit"
	case *ast.BinaryExpr:
		return "bin:" + node.Op.String()
	case *ast.UnaryExpr:
		return "un:" + node.Op.String()
	case *ast.AssignStmt:
		return "assign:" + node.Tok.String()
	case *ast.IncDecStmt:
		return "incdec:" + node.Tok.String()
	case *ast.BranchStmt:
		return "branch:" + node.Tok.String()
	case *ast.IfStmt:
		return "if"
	case *ast.ForStmt:
		return "for"
	case *ast.RangeStmt:
		return "range"
	case *ast.SwitchStmt, *ast.TypeSwitchStmt:
		return "switch"
	case *ast.SelectStmt:
		return "select"
	case *ast.CaseClause, *ast.CommClause:
		return "case"
	case *ast.CallExpr:
		return "call"
	case *ast.DeferStmt:
		return "defer"
	case *ast.GoStmt:
		return "go"
	case *ast.ReturnStmt:
		return "return"
	case *ast.SelectorExpr:
		return "sel"
	case *ast.IndexExpr:
		return "index"
	case *ast.SliceExpr:
		return "slice"
	case *ast.StarExpr:
		return "star"
	case *ast.CompositeLit:
		return "composite"
	case *ast.FuncLit:
		return "funclit"
	case *ast.TypeAssertExpr:
		return "assert"
	default:
		_ = token.ILLEGAL
		return ""
	}
}
