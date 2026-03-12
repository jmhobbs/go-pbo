package main

import (
	"context"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v4"
)

// these variables are overwritten by gorelease at build
var (
	version string = "dev"
	date    string = ""
)

var versionCmd = &ff.Command{
	Name:      "version",
	Usage:     "pbotool version",
	ShortHelp: "Print version and build date information",
	Exec: func(ctx context.Context, args []string) error {
		fmt.Fprintf(os.Stderr, "version: %s\n", version)
		fmt.Fprintf(os.Stderr, "  built: %s\n", date)
		return nil
	},
}
