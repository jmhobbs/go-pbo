package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmhobbs/go-pbo"
	"github.com/peterbourgon/ff/v4"
)

func packCmd() *ff.Command {
	flags := ff.NewFlagSet("pack")
	properties := flags.StringSet('p', "property", "PBO property in the form of key=value. Can be repeated.")
	recursive := flags.Bool('r', "recursive", "Recursively add files from directories.")

	return &ff.Command{
		Name:      "pack",
		Usage:     "pbotool pack [FLAGS] <file.pbo> <input file>...",
		ShortHelp: "Pack files into a PBO",
		Flags:     flags,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return errors.New("missing output file and input file arguments")
			}
			if len(args) == 1 {
				return errors.New("at least one input file is required")
			}

			builder := pbo.NewBuilder()

			if properties != nil {
				for _, property := range *properties {
					split := strings.SplitN(property, "=", 2)
					if len(split) != 2 {
						return fmt.Errorf("invalid property format: %q", property)
					}
					if split[0] == "" || split[1] == "" {
						return errors.New("property key and value cannot be empty")
					}

					builder.SetProperty(split[0], split[1])
				}
			}

			for _, path := range args[1:] {
				finfo, err := os.Stat(path)
				if err != nil {
					return err
				}
				if finfo.IsDir() {
					if *recursive {
						err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
							if err != nil {
								return err
							}
							if d.IsDir() {
								return nil
							}
							fmt.Printf("+ %s\n", path)
							builder.AddFileFromPath(path)
							return nil
						})
						if err != nil {
							return err
						}
					} else {
						return fmt.Errorf("input path is a directory: %q", path)
					}
				} else {
					fmt.Printf("+ %s\n", path)
					builder.AddFileFromPath(path)
				}
			}

			f, err := os.Create(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			return builder.Build(f)
		},
	}
}
