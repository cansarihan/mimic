package srcwalk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkFindsFunctionsAndMethods(t *testing.T) {
	dir := t.TempDir()
	src := `package demo
func Alpha() int { return 1 }
type T struct{}
func (t *T) Beta() {}
func noBodyDecl()
`
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	funcs, err := Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, f := range funcs {
		names[f.Name] = true
	}
	if !names["Alpha"] || !names["(*T).Beta"] {
		t.Errorf("missing functions, got %v", names)
	}
	if names["noBodyDecl"] {
		t.Error("body-less declaration should be skipped")
	}
}

func TestWalkSkipsVendor(t *testing.T) {
	dir := t.TempDir()
	vendor := filepath.Join(dir, "vendor", "x")
	if err := os.MkdirAll(vendor, 0o750); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(vendor, "v.go"), []byte("package x\nfunc Vendored() {}\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc Main() {}\n"), 0o600)
	funcs, err := Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range funcs {
		if f.Name == "Vendored" {
			t.Error("vendored code should be skipped")
		}
	}
}
