package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmhobbs/go-pbo"
	"github.com/peterbourgon/ff/v4"
)

func unpackCmd() *ff.Command {
	return &ff.Command{
		Name:      "unpack",
		Usage:     "pbotool unpack <file.pbo> <output directory>",
		ShortHelp: "Unpack all files from a PBO",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 2 {
				return errors.New("missing file and output directory arguments")
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

			finfo, err := os.Stat(args[1])
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
			if finfo != nil {
				if !finfo.IsDir() {
					return errors.New("output path exists but is not a directory")
				}
			} else {
				if err := os.MkdirAll(args[1], 0755); err != nil {
					return err
				}
			}

			for _, file := range bank.Files {
				log.Printf("Unpacking %s\n", file.Filename)
				var dir string
				// convert windows paths to platform native paths
				segments := strings.Split(file.Filename, "\\")
				filename := segments[len(segments)-1]
				if len(segments) == 1 {
					dir = "."
				} else {
					dir = filepath.Join(segments[:len(segments)-1]...)
				}

				err := os.MkdirAll(filepath.Join(args[1], dir), 0755)
				if err != nil {
					panic(err)
				}
				f, err := os.Create(filepath.Join(args[1], dir, filename))
				if err != nil {
					panic(err)
				}
				defer f.Close()

				reader, err := file.Reader()
				if err != nil {
					return err
				}

				_, err = io.Copy(f, reader)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}
