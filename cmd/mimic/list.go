package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cansarihan/mimic/internal/rules"
)

func runList(args []string) error {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RULE\tDESCRIPTION")
	for _, m := range rules.Registry() {
		fmt.Fprintf(tw, "%s\t%s\n", m.Name(), m.Describe())
	}
	return tw.Flush()
}
