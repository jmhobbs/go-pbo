package io

import (
	"errors"
	gio "io"
)

// io.SectionReader is _almost_ what I want, but it requires io.ReaderAt
// which isn't ideal. So we do a bunch of gymnastics until I find something
// else that works
type limitedReadSeeker struct {
	src              gio.ReadSeeker
	origin           int64
	limit            int64
	relativePosition int64
}

func (l *limitedReadSeeker) Read(p []byte) (int, error) {
	if l.relativePosition >= l.limit {
		return 0, gio.EOF
	}

	if l.relativePosition+int64(len(p)) > l.limit {
		buf := make([]byte, l.limit-l.relativePosition)
		n, err := l.src.Read(buf)
		l.relativePosition += int64(n)
		copy(p, buf[:n])
		return n, err
	}

	n, err := l.src.Read(p)
	l.relativePosition += int64(n)
	return n, err
}

func (l *limitedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var (
		nOffset int64
		err     error
	)
	switch whence {
	case gio.SeekStart:
		l.relativePosition = offset
		if l.relativePosition < 0 {
			return 0, errors.New("error: seek before start")
		}
		nOffset, err = l.src.Seek(l.origin+offset, gio.SeekStart)
	case gio.SeekCurrent:
		l.relativePosition += offset
		if l.relativePosition < 0 {
			return 0, errors.New("error: seek before start")
		}
		nOffset, err = l.src.Seek(l.origin+l.relativePosition, gio.SeekStart)
	case gio.SeekEnd:
		l.relativePosition = l.limit + offset
		if l.relativePosition < 0 {
			return 0, errors.New("error: seek before start")
		}
		nOffset, err = l.src.Seek(l.origin+l.relativePosition, gio.SeekStart)
	default:
		return 0, errors.New("error: invalid seek whence")
	}

	if err != nil {
		return 0, err
	}
	return max(nOffset-l.origin, 0), nil
}

func LimitedReadSeeker(in gio.ReadSeeker, origin, limit int64) (gio.ReadSeeker, error) {
	// ensure we start at the origin
	_, err := in.Seek(origin, gio.SeekStart)
	if err != nil {
		return nil, err
	}
	return &limitedReadSeeker{
		src:              in,
		origin:           origin,
		limit:            limit,
		relativePosition: 0,
	}, nil
}
