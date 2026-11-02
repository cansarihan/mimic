package rules

import (
	"go/ast"

	"github.com/cansarihan/mimic/internal/srcwalk"
)

func init() { register(lockWithoutUnlock{}) }

// lockWithoutUnlock flags a function that acquires a mutex but never releases it.
// It is a heuristic over call names — Lock paired with Unlock, RLock with
// RUnlock — so a function that locks and relies on another to unlock is a known
// false positive, but an outright missing Unlock is a real deadlock risk.
type lockWithoutUnlock struct{}

func (lockWithoutUnlock) Name() string { return "lock-without-unlock" }

func (lockWithoutUnlock) Describe() string {
	return "a mutex is locked but never unlocked in the same function — a likely deadlock"
}

func (l lockWithoutUnlock) Match(fn srcwalk.Func) []Finding {
	calls := map[string]ast.Node{}
	ast.Inspect(fn.Decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if _, seen := calls[sel.Sel.Name]; !seen {
				calls[sel.Sel.Name] = call
			}
		}
		return true
	})

	var out []Finding
	for lock, unlock := range map[string]string{"Lock": "Unlock", "RLock": "RUnlock"} {
		if node, locked := calls[lock]; locked {
			if _, unlocked := calls[unlock]; !unlocked {
				file, line := pos(fn, node)
				out = append(out, Finding{
					Rule: l.Name(), Severity: "error", Func: fn.Name, File: file, Line: line,
					Message: fn.Name + " calls " + lock + " but never " + unlock,
				})
			}
		}
	}
	return out
}
