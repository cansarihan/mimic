package fingerprint

import (
	"go/ast"
	"sort"
)

const shingleSize = 4

// Print is a function's structural fingerprint: the full token stream, a set of
// overlapping k-gram shingles for similarity, and an exact hash of the whole
// structure for detecting identical shapes.
type Print struct {
	Tokens   []string
	Shingles map[uint64]struct{}
	Hash     uint64
	Size     int
}

// New builds a fingerprint from a function declaration.
func New(decl *ast.FuncDecl) Print {
	tokens := Tokens(decl)
	return Print{
		Tokens:   tokens,
		Shingles: shingles(tokens),
		Hash:     hashTokens(tokens),
		Size:     len(tokens),
	}
}

// Similarity is the Jaccard similarity of two fingerprints' shingle sets, from 0
// (nothing in common) to 1 (identical structure). Functions too small to form a
// shingle fall back to an exact-hash comparison.
func Similarity(a, b Print) float64 {
	if len(a.Shingles) == 0 || len(b.Shingles) == 0 {
		if a.Hash == b.Hash && a.Size > 0 {
			return 1
		}
		return 0
	}
	inter := 0
	small, large := a.Shingles, b.Shingles
	if len(small) > len(large) {
		small, large = large, small
	}
	for h := range small {
		if _, ok := large[h]; ok {
			inter++
		}
	}
	union := len(a.Shingles) + len(b.Shingles) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func shingles(tokens []string) map[uint64]struct{} {
	set := make(map[uint64]struct{})
	if len(tokens) < shingleSize {
		if len(tokens) > 0 {
			set[hashTokens(tokens)] = struct{}{}
		}
		return set
	}
	for i := 0; i+shingleSize <= len(tokens); i++ {
		set[hashTokens(tokens[i:i+shingleSize])] = struct{}{}
	}
	return set
}

func hashTokens(tokens []string) uint64 {
	var h uint64 = 1469598103934665603
	for _, t := range tokens {
		for i := 0; i < len(t); i++ {
			h ^= uint64(t[i])
			h *= 1099511628211
		}
		h ^= 0x2c
		h *= 1099511628211
	}
	return h
}

// SortedTokens returns the unique tokens a fingerprint uses, sorted — a compact
// way to show what a function is made of.
func (p Print) SortedTokens() []string {
	seen := map[string]struct{}{}
	for _, t := range p.Tokens {
		seen[t] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
