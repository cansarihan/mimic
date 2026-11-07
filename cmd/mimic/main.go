// Command mimic finds the variants of a known bug: it runs a set of structural
// rules over Go source, and finds the functions that structurally resemble a
// reference you point it at — on the command line or behind a panel.
package main

import (
	"errors"
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "mimic: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return usage()
	}
	switch os.Args[1] {
	case "rules":
		return runRules(os.Args[2:])
	case "like":
		return runLike(os.Args[2:])
	case "list":
		return runList(os.Args[2:])
	case "version":
		fmt.Printf("mimic %s (commit %s, built %s)\n", version, commit, date)
		return nil
	default:
		return usage()
	}
}

func usage() error {
	fmt.Fprintln(os.Stderr, "usage: mimic <rules|like|list|serve|version> [args]")
	fmt.Fprintln(os.Stderr, "  mimic rules ./pkg")
	fmt.Fprintln(os.Stderr, "  mimic like ./pkg validateUser")
	fmt.Fprintln(os.Stderr, "  mimic serve -dir ./pkg -listen 127.0.0.1:8080")
	return errors.New("no command given")
}

func dir(args []string, index int) string {
	if len(args) > index {
		return args[index]
	}
	return "."
}
