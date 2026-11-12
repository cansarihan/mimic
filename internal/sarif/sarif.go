// Package sarif renders a rule scan as a SARIF 2.1.0 document, the format GitHub
// code scanning and other tools ingest, so Mimic's findings can flow into a
// pull request's checks.
package sarif

import (
	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/scan"
)

const (
	schema  = "https://json.schemastore.org/sarif-2.1.0.json"
	version = "2.1.0"
)

// Document is the top level of a SARIF log.
type Document struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name           string `json:"name"`
	InformationURI string `json:"informationUri,omitempty"`
	Rules          []Rule `json:"rules"`
}

type Rule struct {
	ID               string `json:"id"`
	ShortDescription Text   `json:"shortDescription"`
}

type Result struct {
	RuleID    string     `json:"ruleId"`
	Level     string     `json:"level"`
	Message   Text       `json:"message"`
	Locations []Location `json:"locations"`
}

type Text struct {
	Text string `json:"text"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine int `json:"startLine"`
}

// Build converts a rule report into a SARIF document, declaring every registered
// rule in the driver and emitting one result per finding.
func Build(report scan.RuleReport, matchers []rules.Matcher) Document {
	driver := Driver{
		Name:           "mimic",
		InformationURI: "https://github.com/cansarihan/mimic",
	}
	for _, m := range matchers {
		driver.Rules = append(driver.Rules, Rule{ID: m.Name(), ShortDescription: Text{Text: m.Describe()}})
	}

	results := make([]Result, 0, len(report.Findings))
	for _, f := range report.Findings {
		results = append(results, Result{
			RuleID:  f.Rule,
			Level:   level(f.Severity),
			Message: Text{Text: f.Message},
			Locations: []Location{{
				PhysicalLocation: PhysicalLocation{
					ArtifactLocation: ArtifactLocation{URI: f.File},
					Region:           Region{StartLine: f.Line},
				},
			}},
		})
	}

	return Document{
		Schema:  schema,
		Version: version,
		Runs:    []Run{{Tool: Tool{Driver: driver}, Results: results}},
	}
}

// level maps a Mimic severity onto a SARIF level.
func level(severity string) string {
	switch severity {
	case "error":
		return "error"
	case "warning":
		return "warning"
	default:
		return "note"
	}
}
