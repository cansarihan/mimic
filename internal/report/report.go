// Package report renders scan results for the terminal or as JSON.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/cansarihan/mimic/internal/scan"
)

// JSON writes any value as indented JSON.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// RuleTable renders a rule scan: a header tally and one row per finding.
func RuleTable(w io.Writer, report scan.RuleReport) {
	fmt.Fprintf(w, "%s\n%d functions · %d finding(s)\n\n", report.Root, report.Functions, len(report.Findings))
	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "no findings")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SEVERITY\tRULE\tLOCATION\tFUNC\tMESSAGE")
	for _, f := range report.Findings {
		fmt.Fprintf(tw, "%s\t%s\t%s:%d\t%s\t%s\n", f.Severity, f.Rule, short(f.File), f.Line, f.Func, f.Message)
	}
	_ = tw.Flush()
}

// LikeTable renders a sibling scan ranked by similarity.
func LikeTable(w io.Writer, report scan.LikeReport) {
	fmt.Fprintf(w, "reference %s\n%d functions scanned · %d match(es) at ≥ %.0f%%\n\n",
		report.Reference, report.Functions, len(report.Matches), report.Threshold*100)
	if len(report.Matches) == 0 {
		fmt.Fprintln(w, "no siblings above the threshold")
		return
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SIMILARITY\tLOCATION\tFUNC")
	for _, m := range report.Matches {
		fmt.Fprintf(tw, "%.0f%%\t%s:%d\t%s\n", m.Similarity*100, short(m.File), m.Line, m.Func)
	}
	_ = tw.Flush()
}

func short(path string) string {
	const max = 40
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}
