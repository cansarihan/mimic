package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/scan"
)

func TestRuleTable(t *testing.T) {
	var buf bytes.Buffer
	RuleTable(&buf, scan.RuleReport{
		Root: "./pkg", Functions: 10,
		Findings: []rules.Finding{{Rule: "defer-in-loop", Severity: "warning", Func: "f", File: "a.go", Line: 12, Message: "x"}},
	})
	out := buf.String()
	for _, want := range []string{"defer-in-loop", "a.go:12", "1 finding", "SEVERITY"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
}

func TestLikeTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	LikeTable(&buf, scan.LikeReport{Reference: "f", Functions: 5, Threshold: 0.6})
	if !strings.Contains(buf.String(), "no siblings") {
		t.Errorf("empty like table should say so:\n%s", buf.String())
	}
}

func TestClusterTable(t *testing.T) {
	var buf bytes.Buffer
	ClusterTable(&buf, scan.ClusterReport{
		Root: "./pkg", Functions: 4, Threshold: 0.8,
		Clusters: []scan.Cluster{{ID: 1, Size: 2, Members: []scan.ClusterMember{
			{Func: "validateUser", File: "auth.go", Line: 20},
			{Func: "validateAdmin", File: "auth.go", Line: 33},
		}}},
	})
	out := buf.String()
	for _, want := range []string{"1 clone cluster", "cluster 1 · 2 members", "validateAdmin", "auth.go:33"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q\n%s", want, out)
		}
	}
}
