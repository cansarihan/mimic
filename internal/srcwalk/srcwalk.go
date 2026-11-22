// Package srcwalk parses the Go source under a directory and yields every
// function it finds, with the position information needed to point a human back
// at the code. It is the front end every analysis in Mimic shares.
package srcwalk

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Func is one function or method recovered from the source.
type Func struct {
	Package string
	Name    string
	File    string
	Line    int
	Decl    *ast.FuncDecl
	Fset    *token.FileSet
}

// Walk parses every non-vendored Go file under root and returns its functions.
// Generated files (marked `Code generated ... DO NOT EDIT.`) are skipped. Test
// files are included; a function with no body (an assembly stub) is skipped
// because there is nothing to analyze.
func Walk(root string) ([]Func, error) {
	fset := token.NewFileSet()
	var funcs []Func

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(d.Name()) && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution|parser.ParseComments)
		if parseErr != nil {
			return nil // a file that does not parse is skipped, not fatal
		}
		if ast.IsGenerated(file) {
			return nil // generated code is not where variants are fixed
		}
		pkg := file.Name.Name
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			funcs = append(funcs, Func{
				Package: pkg,
				Name:    funcName(fn),
				File:    path,
				Line:    fset.Position(fn.Pos()).Line,
				Decl:    fn,
				Fset:    fset,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(funcs, func(i, j int) bool {
		if funcs[i].File != funcs[j].File {
			return funcs[i].File < funcs[j].File
		}
		return funcs[i].Line < funcs[j].Line
	})
	return funcs, nil
}

func skipDir(name string) bool {
	return name == "vendor" || name == "testdata" || (strings.HasPrefix(name, ".") && name != ".")
}

func funcName(fn *ast.FuncDecl) string {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		return receiverName(fn.Recv.List[0].Type) + "." + fn.Name.Name
	}
	return fn.Name.Name
}

func receiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return "(*" + receiverName(t.X) + ")"
	case *ast.Ident:
		return t.Name
	default:
		return "?"
	}
}
