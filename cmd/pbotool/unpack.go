package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmhobbs/go-pbo"
	"github.com/jmhobbs/go-raP"
	"github.com/jmhobbs/go-raP/printer"
	"github.com/peterbourgon/ff/v4"
)

func unpackCmd() *ff.Command {
	flags := ff.NewFlagSet("unpack")
	unrap := flags.Bool('u', "unrap", "De-binarize raP files (i.e. config.bin -> config.cpp)")

	return &ff.Command{
		Name:      "unpack",
		Usage:     "pbotool unpack [flags] <file.pbo> <output directory>",
		Flags:     flags,
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
				if file.DataSize == 0 {
					log.Println("  Empty file, skipping")
					continue
				}

				reader, err := file.ReadSeeker()
				if err != nil {
					return err
				}

				var shouldUnrap bool
				if *unrap {
					shouldUnrap, err = isBinarizedFile(reader)
					if err != nil {
						return err
					}
					_, err = reader.Seek(0, io.SeekStart) // Reset reader for actual reading
					if err != nil {
						return err
					}
				}

				var dir string
				// convert windows paths to platform native paths
				segments := strings.Split(file.Filename, "\\")
				filename := segments[len(segments)-1]
				if len(segments) == 1 {
					dir = "."
				} else {
					dir = filepath.Join(segments[:len(segments)-1]...)
				}

				if shouldUnrap {
					filename = strings.TrimSuffix(filename, ".bin") + ".cpp"
				}

				err = os.MkdirAll(filepath.Join(args[1], dir), 0755)
				if err != nil {
					panic(err)
				}
				f, err := os.Create(filepath.Join(args[1], dir, filename))
				if err != nil {
					panic(err)
				}
				defer f.Close()

				if shouldUnrap {
					root, err := raP.Decode(reader)
					if err != nil {
						return err
					}
					err = printer.New().File(f, root)
					if err != nil {
						return err
					}
				} else {
					_, err = io.Copy(f, reader)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}
}

func isBinarizedFile(in io.ReadSeeker) (bool, error) {
	fingerprint := make([]byte, 4)
	_, err := in.Read(fingerprint)
	if err != nil {
		return false, err
	}

	return bytes.Equal([]byte{0, 'r', 'a', 'P'}, fingerprint), nil
}
