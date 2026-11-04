// Package scan ties the front end, the rules and the fingerprints together into
// the two analyses Mimic offers: running the built-in rules over a tree, and
// finding the structural siblings of one reference function.
package scan

import (
	"sort"

	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/srcwalk"
)

// RuleReport is the result of running the rule set over a source tree.
type RuleReport struct {
	Root       string          `json:"root"`
	Functions  int             `json:"functions"`
	Findings   []rules.Finding `json:"findings"`
	Counts     map[string]int  `json:"counts"`
	BySeverity map[string]int  `json:"by_severity"`
}

// Rules runs every matcher over every function under root.
func Rules(root string, matchers []rules.Matcher) (RuleReport, error) {
	funcs, err := srcwalk.Walk(root)
	if err != nil {
		return RuleReport{}, err
	}
	report := RuleReport{
		Root:       root,
		Functions:  len(funcs),
		Counts:     map[string]int{},
		BySeverity: map[string]int{},
	}
	for _, fn := range funcs {
		for _, m := range matchers {
			for _, finding := range m.Match(fn) {
				report.Findings = append(report.Findings, finding)
				report.Counts[finding.Rule]++
				report.BySeverity[finding.Severity]++
			}
		}
	}
	sort.SliceStable(report.Findings, func(i, j int) bool {
		if report.Findings[i].File != report.Findings[j].File {
			return report.Findings[i].File < report.Findings[j].File
		}
		return report.Findings[i].Line < report.Findings[j].Line
	})
	return report, nil
}
