package main

import (
	"flag"
	"os"

	"github.com/cansarihan/mimic/internal/report"
	"github.com/cansarihan/mimic/internal/scan"
)

func runClusters(args []string) error {
	flags := flag.NewFlagSet("clusters", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print JSON instead of a listing")
	min := flags.Float64("min", 0.75, "minimum similarity for two functions to cluster (0..1)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	report_, err := scan.Clusters(dir(flags.Args(), 0), *min)
	if err != nil {
		return err
	}
	if *asJSON {
		return report.JSON(os.Stdout, report_)
	}
	report.ClusterTable(os.Stdout, report_)
	return nil
}
