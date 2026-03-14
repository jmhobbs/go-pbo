package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/jmhobbs/go-pbo"
	"github.com/peterbourgon/ff/v4"
)

var grepableFileExtensions = []string{".cpp", ".c", ".txt", ".xml", ".json", ".yml"}

func grepCmd() *ff.Command {
	flags := ff.NewFlagSet("grep")
	caseInsensitive := flags.Bool('i', "ignore-case", "Perform case-insensitive matching.")

	return &ff.Command{
		Name:      "grep",
		Usage:     "pbotool grep <file.pbo> <pattern>",
		ShortHelp: "Search non-binary files in a PBO (RE2 syntax)",
		Flags:     flags,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 2 {
				return errors.New("missing file and pattern arguments")
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

			var pattern *regexp.Regexp

			if *caseInsensitive {
				pattern, err = regexp.Compile("(?i)" + args[1])
			} else {
				pattern, err = regexp.Compile(args[1])
			}
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}

			for _, file := range bank.Files {
				ext := filepath.Ext(file.Filename)
				if slices.Contains(grepableFileExtensions, ext) {
					reader, err := file.Reader()
					if err != nil {
						return err
					}
					scanner := bufio.NewScanner(reader)
					line := 0
					for scanner.Scan() {
						line += 1
						if pattern.Match(scanner.Bytes()) {
							fmt.Printf("%s:%d:%s\n", file.Filename, line, scanner.Text())
						}
					}
					if err := scanner.Err(); err != nil {
						fmt.Fprintln(os.Stderr, "reading standard input:", err)
					}
				}
			}

			return nil
		},
	}
}
