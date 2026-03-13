package main

import (
	"context"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	rootCmd := &ff.Command{
		Name:  "pbotool",
		Usage: "pbotool SUBCOMMAND ...",
	}
	rootCmd.Exec = func(ctx context.Context, args []string) error {
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
		return nil
	}

	rootCmd.Subcommands = append(rootCmd.Subcommands, inspectCmd())
	rootCmd.Subcommands = append(rootCmd.Subcommands, grepCmd())
	rootCmd.Subcommands = append(rootCmd.Subcommands, unpackCmd())
	rootCmd.Subcommands = append(rootCmd.Subcommands, packCmd())
	rootCmd.Subcommands = append(rootCmd.Subcommands, versionCmd)

	if err := rootCmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Command(rootCmd))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
