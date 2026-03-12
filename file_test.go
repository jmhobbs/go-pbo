package pbo_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/jmhobbs/go-pbo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_File_Reader(t *testing.T) {
	t.Run("reads file content", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// File header: "test.txt"
			't', 'e', 's', 't', '.', 't', 'x', 't', 0,
			0, 0, 0, 0, // type: Dummy
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			5, 0, 0, 0, // data size: 5
			// End header
			0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			// data
			'h', 'e', 'l', 'l', 'o',
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 1)

		r, err := bank.Files[0].Reader()
		require.NoError(t, err)

		content, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, []byte("hello"), content)
	})

	t.Run("reads correct content for each file in multi-file PBO", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// Header for "a.txt"
			'a', '.', 't', 'x', 't', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			3, 0, 0, 0, // data size: 3
			// Header for "b.txt"
			'b', '.', 't', 'x', 't', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			4, 0, 0, 0, // data size: 4
			// End header
			0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			// data
			'a', 'b', 'c',
			'd', 'e', 'f', 'g',
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 2)

		r0, err := bank.Files[0].Reader()
		require.NoError(t, err)
		c0, err := io.ReadAll(r0)
		require.NoError(t, err)
		assert.Equal(t, []byte("abc"), c0)

		r1, err := bank.Files[1].Reader()
		require.NoError(t, err)
		c1, err := io.ReadAll(r1)
		require.NoError(t, err)
		assert.Equal(t, []byte("defg"), c1)
	})

	t.Run("reader can be called multiple times", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			't', 'e', 's', 't', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			3, 0, 0, 0,
			0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			'x', 'y', 'z',
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 1)

		for i := 0; i < 3; i++ {
			r, err := bank.Files[0].Reader()
			require.NoError(t, err)
			content, err := io.ReadAll(r)
			require.NoError(t, err)
			assert.Equal(t, []byte("xyz"), content)
		}
	})
}
