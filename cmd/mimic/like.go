package main

import (
	"errors"
	"flag"
	"os"

	"github.com/cansarihan/mimic/internal/report"
	"github.com/cansarihan/mimic/internal/scan"
)

func runLike(args []string) error {
	flags := flag.NewFlagSet("like", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "print JSON instead of a table")
	min := flags.Float64("min", 0.6, "minimum similarity (0..1)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() < 2 {
		return errors.New("like needs a directory and a reference: mimic like ./pkg FuncName")
	}

	report_, err := scan.Like(flags.Arg(0), flags.Arg(1), *min)
	if err != nil {
		return err
	}
	if *asJSON {
		return report.JSON(os.Stdout, report_)
	}
	report.LikeTable(os.Stdout, report_)
	return nil
}
