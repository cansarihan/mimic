package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cansarihan/mimic/internal/rules"
)

func writeSrc(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte("package p\n"+src), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRulesFindsBugs(t *testing.T) {
	dir := writeSrc(t, `
import "os"
func leak(paths []string) {
	for _, p := range paths {
		f, _ := os.Open(p)
		defer f.Close()
		_ = f
	}
}
func lost(xs []int) { append(xs, 1) }
`)
	report, err := Rules(dir, rules.Registry())
	if err != nil {
		t.Fatal(err)
	}
	if report.Counts["defer-in-loop"] != 1 {
		t.Errorf("defer-in-loop count = %d, want 1", report.Counts["defer-in-loop"])
	}
	if report.Counts["discarded-append"] != 1 {
		t.Errorf("discarded-append count = %d, want 1", report.Counts["discarded-append"])
	}
	if report.BySeverity["error"] < 1 {
		t.Errorf("expected at least one error-severity finding")
	}
}

func TestLikeRanksSibling(t *testing.T) {
	dir := writeSrc(t, `
func sumA(xs []int) int { total := 0; for _, x := range xs { total += x }; return total }
func sumB(ys []int) int { acc := 0; for _, y := range ys { acc += y }; return acc }
func greet(s string) string { return "hi " + s }
`)
	report, err := Like(dir, "sumA", 0.4)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Matches) == 0 {
		t.Fatal("expected at least one sibling match")
	}
	if report.Matches[0].Func != "sumB" {
		t.Errorf("top match = %q, want sumB", report.Matches[0].Func)
	}
	for _, m := range report.Matches {
		if m.Func == "greet" {
			t.Error("unrelated function should be below threshold")
		}
	}
}

func TestLikeMissingReference(t *testing.T) {
	dir := writeSrc(t, `func f() {}`)
	if _, err := Like(dir, "nope", 0.5); err == nil {
		t.Fatal("expected an error for a missing reference")
	}
}
