package scan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cansarihan/mimic/internal/fingerprint"
	"github.com/cansarihan/mimic/internal/srcwalk"
)

// Match is one function that resembles the reference, with its similarity.
type Match struct {
	Func       string  `json:"func"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Similarity float64 `json:"similarity"`
}

// LikeReport is the ranked set of a reference function's structural siblings.
type LikeReport struct {
	Root      string  `json:"root"`
	Reference string  `json:"reference"`
	Functions int     `json:"functions"`
	Threshold float64 `json:"threshold"`
	Matches   []Match `json:"matches"`
}

// Like fingerprints the reference function named ref (matched by name, optionally
// qualified by a file substring as "file.go:Name") and ranks every other
// function by structural similarity, keeping those at or above threshold.
func Like(root, ref string, threshold float64) (LikeReport, error) {
	funcs, err := srcwalk.Walk(root)
	if err != nil {
		return LikeReport{}, err
	}

	fileHint, name := splitRef(ref)
	var reference *srcwalk.Func
	for i := range funcs {
		if funcs[i].Name == name && (fileHint == "" || strings.Contains(funcs[i].File, fileHint)) {
			reference = &funcs[i]
			break
		}
	}
	if reference == nil {
		return LikeReport{}, fmt.Errorf("reference function %q not found under %s", ref, root)
	}

	refPrint := fingerprint.New(reference.Decl)
	report := LikeReport{Root: root, Reference: reference.Name, Functions: len(funcs), Threshold: threshold}
	for i := range funcs {
		if &funcs[i] == reference {
			continue
		}
		sim := fingerprint.Similarity(refPrint, fingerprint.New(funcs[i].Decl))
		if sim >= threshold {
			report.Matches = append(report.Matches, Match{
				Func: funcs[i].Name, File: funcs[i].File, Line: funcs[i].Line, Similarity: sim,
			})
		}
	}
	sort.SliceStable(report.Matches, func(i, j int) bool {
		return report.Matches[i].Similarity > report.Matches[j].Similarity
	})
	return report, nil
}

func splitRef(ref string) (fileHint, name string) {
	if idx := strings.LastIndex(ref, ":"); idx >= 0 {
		return ref[:idx], ref[idx+1:]
	}
	return "", ref
}
