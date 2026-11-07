package main

import (
	"flag"
	"os"
	"strings"

	"github.com/cansarihan/mimic/internal/report"
	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/scan"
)

func runRules(args []string) error {
	flags := flag.NewFlagSet("rules", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print JSON instead of a table")
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
	if *asJSON {
		return report.JSON(os.Stdout, report_)
	}
	report.RuleTable(os.Stdout, report_)
	return nil
}
