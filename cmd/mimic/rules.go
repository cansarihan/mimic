package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cansarihan/mimic/internal/report"
	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/sarif"
	"github.com/cansarihan/mimic/internal/scan"
)

func runRules(args []string) error {
	flags := flag.NewFlagSet("rules", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print JSON instead of a table")
	sarifOut := flags.Bool("sarif", false, "print a SARIF 2.1.0 document")
	failOn := flags.String("fail-on", "", "exit non-zero when a finding is at or above this severity (error, warning, info)")
	only := flags.String("rule", "", "only run the named rule")
	if err := flags.Parse(args); err != nil {
		return err
	}

	matchers := rules.Registry()
	if *only != "" {
		filtered := matchers[:0]
		for _, m := range matchers {
			if strings.EqualFold(m.Name(), *only) {
				filtered = append(filtered, m)
			}
		}
		matchers = filtered
	}

	report_, err := scan.Rules(dir(flags.Args(), 0), matchers)
	if err != nil {
		return err
	}
	switch {
	case *sarifOut:
		err = report.JSON(os.Stdout, sarif.Build(report_, rules.Registry()))
	case *asJSON:
		err = report.JSON(os.Stdout, report_)
	default:
		report.RuleTable(os.Stdout, report_)
	}
	if err != nil {
		return err
	}
	if n := report_.AtOrAbove(*failOn); n > 0 {
		return fmt.Errorf("%d finding(s) at or above %s", n, *failOn)
	}
	return nil
}
