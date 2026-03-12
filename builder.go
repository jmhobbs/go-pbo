package pbo

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type fileWithReader struct {
	in io.Reader
	fileMetadata
}
type fileMetadata struct {
	path      string
	timestamp time.Time
	size      uint32
}

type builder struct {
	properties      map[string]string
	files           []fileMetadata
	filesWithReader []fileWithReader
}

func NewBuilder() *builder {
	return &builder{properties: make(map[string]string)}
}

func (b *builder) AddFile(in io.Reader, path string, timestamp time.Time, size uint32) *builder {
	b.filesWithReader = append(b.filesWithReader, fileWithReader{in, fileMetadata{path, timestamp, size}})
	return b
}

func (b *builder) AddFileFromPath(path string) *builder {
	b.files = append(b.files, fileMetadata{path: path})
	return b
}

func (b *builder) SetProperty(key, value string) *builder {
	b.properties[key] = value
	return b
}

func (b *builder) hydrateFiles() error {
	for i, file := range b.files {
		finfo, err := os.Stat(file.path)
		if err != nil {
			return fmt.Errorf("error with input file %q: %w", file.path, err)
		}
		if finfo.IsDir() {
			return fmt.Errorf("input file %q is a directory", file.path)
		}
		file.timestamp = finfo.ModTime()
		file.size = uint32(finfo.Size())
		b.files[i] = file
	}
	return nil
}

func writeHeader(out io.Writer, path string, hdr header) error {
	var err error

	if len(path) > 0 {
		_, err = out.Write([]byte(path))
		if err != nil {
			return err
		}
	}
	_, err = out.Write([]byte{0})
	if err != nil {
		return err
	}

	return binary.Write(out, binary.LittleEndian, hdr)
}

func writeProperties(out io.Writer, properties map[string]string) error {
	// Properties
	for key, value := range properties {
		err := binary.Write(out, binary.LittleEndian, []byte(key))
		if err != nil {
			return err
		}
		_, err = out.Write([]byte{0})
		if err != nil {
			return err
		}
		err = binary.Write(out, binary.LittleEndian, []byte(value))
		if err != nil {
			return err
		}
		_, err = out.Write([]byte{0})
		if err != nil {
			return err
		}
	}
	_, err := out.Write([]byte{0})

	return err
}

func writeFileHeader(out io.Writer, file fileMetadata) error {
	// normalize paths
	split := strings.Split(file.path, string(os.PathSeparator))
	path := strings.Join(split, `\`)

	return writeHeader(
		out,
		path,
		header{
			Type:         0,
			OriginalSize: 0,
			Offset:       0,
			Timestamp:    uint32(file.timestamp.Unix()),
			DataSize:     file.size,
		},
	)
}

func (b *builder) Build(out io.Writer) error {
	// fill out timestamp and data sizes for non-reader files
	err := b.hydrateFiles()
	if err != nil {
		return err
	}

	// Vers header
	err = writeHeader(out, "", header{Type: Vers})
	if err != nil {
		return err
	}

	err = writeProperties(out, b.properties)
	if err != nil {
		return err
	}

	// File headers
	for _, file := range b.files {
		err = writeFileHeader(out, file)
		if err != nil {
			return err
		}
	}
	for _, file := range b.filesWithReader {
		err = writeFileHeader(out, file.fileMetadata)
		if err != nil {
			return err
		}
	}

	// final marker header
	err = writeHeader(out, "", header{Type: 0})
	if err != nil {
		return err
	}

	// data
	for _, file := range b.files {
		f, err := os.Open(file.path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(out, f)
		if err != nil {
			return err
		}
	}

	for _, file := range b.filesWithReader {
		_, err = io.Copy(out, file.in)
		if err != nil {
			return err
		}
	}

	return nil
}
