package fingerprint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseFunc(t *testing.T, src string) *ast.FuncDecl {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "x.go", "package p\n"+src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	return file.Decls[0].(*ast.FuncDecl)
}

func TestRenamingKeepsStructure(t *testing.T) {
	a := New(parseFunc(t, `func f(xs []int) int { total := 0; for _, x := range xs { total += x }; return total }`))
	b := New(parseFunc(t, `func g(ys []int) int { sum := 0; for _, y := range ys { sum += y }; return sum }`))
	if s := Similarity(a, b); s < 0.99 {
		t.Errorf("renamed clone similarity = %v, want ~1", s)
	}
}

func TestDifferentStructureDiffers(t *testing.T) {
	a := New(parseFunc(t, `func f(xs []int) int { total := 0; for _, x := range xs { total += x }; return total }`))
	b := New(parseFunc(t, `func h(s string) bool { return s == "" }`))
	if sim := Similarity(a, b); sim > 0.3 {
		t.Errorf("unrelated functions similarity = %v, want low", sim)
	}
}

func TestEditedSiblingIsSimilarNotIdentical(t *testing.T) {
	a := New(parseFunc(t, `func f(xs []int) int { total := 0; for _, x := range xs { total += x }; return total }`))
	b := New(parseFunc(t, `func f2(xs []int) int { total := 0; for _, x := range xs { if x > 0 { total += x } }; return total }`))
	s := Similarity(a, b)
	if s < 0.3 || s >= 1 {
		t.Errorf("edited sibling similarity = %v, want partial", s)
	}
}
