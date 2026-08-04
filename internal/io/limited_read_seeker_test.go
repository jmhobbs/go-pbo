package io_test

import (
	"bytes"
	"errors"
	gio "io"
	"testing"

	"github.com/jmhobbs/go-pbo/internal/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// source is 20 bytes: index 0-19. Window [5,15) -> "56789ABCDE".
const limitedReadSeekerSource = "0123456789ABCDEFGHIJ"

// eofWithDataReader is an io.ReadSeeker that, like some real-world readers
// (e.g. net.Conn, archive/tar), returns a final non-zero read together with
// a non-nil error instead of returning the data and EOF on separate calls.
type eofWithDataReader struct {
	data []byte
	pos  int
}

func (r *eofWithDataReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, gio.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	var err error
	if r.pos >= len(r.data) {
		err = gio.EOF
	}
	return n, err
}

func (r *eofWithDataReader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case gio.SeekStart:
		newPos = offset
	case gio.SeekCurrent:
		newPos = int64(r.pos) + offset
	case gio.SeekEnd:
		newPos = int64(len(r.data)) + offset
	}
	r.pos = int(newPos)
	return newPos, nil
}

func Test_LimitedReadSeeker(t *testing.T) {
	t.Run("reads the window starting at origin", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		content, err := gio.ReadAll(lrs)
		require.NoError(t, err)
		assert.Equal(t, "56789ABCDE", string(content))
	})

	t.Run("does not read past the limit", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		buf := make([]byte, 100)
		n, err := lrs.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, "56789ABCDE", string(buf[:n]))

		n, err = lrs.Read(buf)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, gio.EOF)
	})

	t.Run("SeekStart positions reads within the window", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		_, err = lrs.Seek(3, gio.SeekStart)
		require.NoError(t, err)

		content, err := gio.ReadAll(lrs)
		require.NoError(t, err)
		assert.Equal(t, "89ABCDE", string(content))
	})

	t.Run("SeekEnd positions reads relative to the end of the window", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		_, err = lrs.Seek(-3, gio.SeekEnd)
		require.NoError(t, err)

		content, err := gio.ReadAll(lrs)
		require.NoError(t, err)
		assert.Equal(t, "CDE", string(content))
	})

	t.Run("SeekCurrent moves relative to the current position after a read", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		buf := make([]byte, 3)
		n, err := lrs.Read(buf) // consumes "567", position now 3
		require.NoError(t, err)
		require.Equal(t, 3, n)

		pos, err := lrs.Seek(2, gio.SeekCurrent) // should land on position 5
		require.NoError(t, err)
		require.EqualValues(t, 5, pos)

		content, err := gio.ReadAll(lrs)
		require.NoError(t, err)
		assert.Equal(t, "ABCDE", string(content))
	})

	t.Run("Seek returns the new relative offset", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		pos, err := lrs.Seek(3, gio.SeekStart)
		require.NoError(t, err)
		assert.EqualValues(t, 3, pos)

		pos, err = lrs.Seek(-2, gio.SeekEnd)
		require.NoError(t, err)
		assert.EqualValues(t, 8, pos)
	})

	t.Run("seeking before the start of the window is an error", func(t *testing.T) {
		src := bytes.NewReader([]byte(limitedReadSeekerSource))
		lrs, err := io.LimitedReadSeeker(src, 5, 10)
		require.NoError(t, err)

		_, err = lrs.Seek(-1, gio.SeekStart)
		assert.Error(t, err)

		_, err = lrs.Seek(0, gio.SeekStart)
		require.NoError(t, err)
		_, err = lrs.Seek(-1, gio.SeekCurrent)
		assert.Error(t, err)

		_, err = lrs.Seek(-11, gio.SeekEnd)
		assert.Error(t, err)
	})

	t.Run("forwards data returned alongside a non-nil error", func(t *testing.T) {
		// The underlying reader returns its final chunk together with
		// gio.EOF in the same call, which is legal per the io.Reader
		// contract (e.g. net.Conn, archive/tar do this).
		src := &eofWithDataReader{data: []byte("0123456789")}
		lrs, err := io.LimitedReadSeeker(src, 0, 10)
		require.NoError(t, err)

		buf := make([]byte, 10)
		n, err := lrs.Read(buf)
		assert.True(t, err == nil || errors.Is(err, gio.EOF))
		assert.Equal(t, "0123456789", string(buf[:n]))
	})
}
