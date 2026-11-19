package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

// matchSource parses one source file and runs a matcher over every function.
func matchSource(t *testing.T, m Matcher, src string) []Finding {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte("package p\n"+src), 0o600); err != nil {
		t.Fatal(err)
	}
	funcs, err := srcwalk.Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []Finding
	for _, fn := range funcs {
		out = append(out, m.Match(fn)...)
	}
	return out
}

func TestDeferInLoop(t *testing.T) {
	hits := matchSource(t, deferInLoop{}, `
import "os"
func bad(paths []string) {
	for _, p := range paths {
		f, _ := os.Open(p)
		defer f.Close()
		_ = f
	}
}
func good(paths []string) {
	for _, p := range paths {
		func() {
			f, _ := os.Open(p)
			defer f.Close()
			_ = f
		}()
	}
}
`)
	if len(hits) != 1 {
		t.Fatalf("want 1 finding (bad only), got %d: %+v", len(hits), hits)
	}
	if hits[0].Func != "bad" {
		t.Errorf("finding in %q, want bad", hits[0].Func)
	}
}

func TestDiscardedAppend(t *testing.T) {
	hits := matchSource(t, discardedAppend{}, `
func bad(xs []int) { _ = append(xs, 1) }
func good(xs []int) []int { return append(xs, 1) }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestSprintfQuery(t *testing.T) {
	hits := matchSource(t, sprintfQuery{}, `
import ("database/sql"; "fmt")
func bad(db *sql.DB, id string) { db.Query(fmt.Sprintf("select * from t where id=%s", id)) }
func good(db *sql.DB, id string) { db.Query("select * from t where id=$1", id) }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestLockWithoutUnlock(t *testing.T) {
	hits := matchSource(t, lockWithoutUnlock{}, `
import "sync"
func bad(mu *sync.Mutex) { mu.Lock(); _ = 1 }
func good(mu *sync.Mutex) { mu.Lock(); defer mu.Unlock() }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestRandSecret(t *testing.T) {
	hits := matchSource(t, randSecret{}, `
import "math/rand"
func generateToken() int { return rand.Intn(1000000) }
func rollDice() int { return rand.Intn(6) }
`)
	if len(hits) != 1 || hits[0].Func != "generateToken" {
		t.Fatalf("want 1 finding in generateToken, got %+v", hits)
	}
}

func TestSwallowedError(t *testing.T) {
	hits := matchSource(t, swallowedError{}, `
import "errors"
func bad() { err := errors.New("x"); if err != nil { } }
func good() error { err := errors.New("x"); if err != nil { return err }; return nil }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestUncheckedAssert(t *testing.T) {
	hits := matchSource(t, uncheckedAssert{}, `
func bad(x any) int { return x.(int) }
func good(x any) int { v, ok := x.(int); if !ok { return 0 }; return v }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestHardcodedCreds(t *testing.T) {
	hits := matchSource(t, hardcodedCreds{}, `
func bad() { apiKey := "sk-abc123secret" ; _ = apiKey }
func good(k string) { apiKey := k ; _ = apiKey }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestRegexpInLoop(t *testing.T) {
	hits := matchSource(t, regexpInLoop{}, `
import "regexp"
func bad(xs []string) { for _, x := range xs { re := regexp.MustCompile("a+"); _ = re.MatchString(x) } }
func good(xs []string) { re := regexp.MustCompile("a+"); for _, x := range xs { _ = re.MatchString(x) } }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestHTTPDefaultClient(t *testing.T) {
	hits := matchSource(t, httpDefaultClient{}, `
import ("net/http"; "time")
func bad() { http.Get("https://x") }
func good() { c := &http.Client{Timeout: time.Second}; c.Get("https://x") }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestTimeAfterInLoop(t *testing.T) {
	hits := matchSource(t, timeAfterInLoop{}, `
import "time"
func bad(ch chan int) { for { select { case <-ch: case <-time.After(time.Second): return } } }
func good(ch chan int) { t := time.NewTimer(time.Second); for { select { case <-ch: case <-t.C: return } } }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad, got %+v", hits)
	}
}

func TestOSExitInLibrary(t *testing.T) {
	hits := matchSource(t, osExitInLibrary{}, `
import "os"
func bad() { os.Exit(1) }
`)
	if len(hits) != 1 || hits[0].Func != "bad" {
		t.Fatalf("want 1 finding in bad (package p), got %+v", hits)
	}
}

func TestRegistryNotEmpty(t *testing.T) {
	if len(Registry()) == 0 {
		t.Fatal("no matchers registered")
	}
}
