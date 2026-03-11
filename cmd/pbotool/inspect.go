package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jmhobbs/go-pbo"
	"github.com/peterbourgon/ff/v4"
)

func inspectCmd() *ff.Command {
	return &ff.Command{
		Name:      "inspect",
		Usage:     "pbotool inspect <file.pbo>",
		ShortHelp: "Inspect the contents of a PBO",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return errors.New("missing file argument")
			}

			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer f.Close()

			bank, err := pbo.Load(f)
			if err != nil {
				return fmt.Errorf("failed to load PBO: %w", err)
			}

			fmt.Println("[Properties]")
			for key, value := range bank.Properties {
				fmt.Printf("- %s: %s\n", key, value)
			}
			fmt.Println("")

			fmt.Println("[Files]")
			for _, file := range bank.Files {
				fmt.Printf("- %s (%d bytes)\n", file.Filename, file.DataSize)
			}

			return nil
		},
	}
}
