package sarif

import (
	"encoding/json"
	"testing"

	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/scan"
)

func TestBuild(t *testing.T) {
	report := scan.RuleReport{
		Root: "./pkg",
		Findings: []rules.Finding{
			{Rule: "defer-in-loop", Severity: "warning", File: "a.go", Line: 12, Message: "x"},
			{Rule: "sprintf-into-query", Severity: "error", File: "b.go", Line: 3, Message: "y"},
		},
	}
	doc := Build(report, rules.Registry())
	if doc.Version != "2.1.0" || len(doc.Runs) != 1 {
		t.Fatalf("bad document shell: %+v", doc)
	}
	run := doc.Runs[0]
	if run.Tool.Driver.Name != "mimic" {
		t.Errorf("driver name = %q", run.Tool.Driver.Name)
	}
	if len(run.Tool.Driver.Rules) != len(rules.Registry()) {
		t.Errorf("declared %d rules, want %d", len(run.Tool.Driver.Rules), len(rules.Registry()))
	}
	if len(run.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(run.Results))
	}
	if run.Results[1].Level != "error" || run.Results[1].Locations[0].PhysicalLocation.Region.StartLine != 3 {
		t.Errorf("result mapping wrong: %+v", run.Results[1])
	}

	// It must round-trip as JSON.
	if _, err := json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
}

func TestLevelMapping(t *testing.T) {
	if level("error") != "error" || level("warning") != "warning" || level("info") != "note" {
		t.Error("severity to SARIF level mapping is wrong")
	}
}
