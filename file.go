package pbo

import (
	"encoding/binary"
	"io"

	iio "github.com/jmhobbs/go-pbo/internal/io"
)

type MimeType uint32

const (
	Vers  MimeType = 0x56657273
	Cprs  MimeType = 0x43707273
	Enco  MimeType = 0x456e6372
	Dummy MimeType = 0x00000000
)

type File struct {
	Filename string
	header

	reader io.ReadSeeker
	offset uint32 // combined offset in the data segment
}

func (f *File) Reader() (io.Reader, error) {
	return f.ReadSeeker()
}

func (f *File) ReadSeeker() (io.ReadSeeker, error) {
	if f.reader == nil || f.offset == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	// TODO: Support compressed files with a wrapper around reader

	return iio.LimitedReadSeeker(f.reader, int64(f.offset), int64(f.DataSize))
}

type header struct {
	Type         MimeType
	OriginalSize uint32
	Offset       uint32
	Timestamp    uint32
	DataSize     uint32
}

func readHeader(in io.Reader) (*File, error) {
	filename, err := readAsciiz(in)
	if err != nil {
		return nil, err
	}

	h := header{}
	err = binary.Read(in, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}

	// TODO: Validate

	return &File{
		Filename: filename,
		header:   h,
	}, nil
}
