package pbo_test

import (
	"bytes"
	"testing"

	"github.com/jmhobbs/go-pbo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Load(t *testing.T) {
	t.Run("minimal PBO with no properties or files", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// end header
			0,          // no filename
			0, 0, 0, 0, // type
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// no data
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.NotNil(t, bank)
		assert.Empty(t, bank.Properties)
		assert.Empty(t, bank.Files)
	})

	t.Run("PBO with properties but no files", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// Vers header
			0,                      // no filename
			0x73, 0x72, 0x65, 0x56, // type: "Vers"
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// properties
			'h', 'e', 'l', 'l', 'o', 0, // key
			'w', 'o', 'r', 'l', 'd', 0, // value
			0, // final empty string delimiter
			// end header
			0,          // no filename
			0, 0, 0, 0, // type
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// no data
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.Len(t, bank.Properties, 1)
		assert.Equal(t, "world", bank.Properties["hello"])
	})

	t.Run("EOF during header read", func(t *testing.T) {
		reader := bytes.NewReader([]byte("incomplete"))
		_, err := pbo.Load(reader)
		require.Error(t, err)
	})

	t.Run("PBO with property key but not value", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// Vers header
			0,                      // no filename
			0x73, 0x72, 0x65, 0x56, // type: "Vers"
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// properties
			'h', 'e', 'l', 'l', 'o', 0, // key
			// no value
			0, // final empty string delimiter
			// end header
			0,          // no filename
			0, 0, 0, 0, // type
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// no data
		})

		_, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.Error(t, err)
	})

	t.Run("PBO with one file", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// File header: "test.txt"
			't', 'e', 's', 't', '.', 't', 'x', 't', 0, // filename
			0, 0, 0, 0, // type: Dummy
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			5, 0, 0, 0, // data size: 5
			// End header
			0,          // no filename
			0, 0, 0, 0, // type
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// data
			'h', 'e', 'l', 'l', 'o',
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 1)
		assert.Equal(t, "test.txt", bank.Files[0].Filename)
		assert.Equal(t, uint32(5), bank.Files[0].DataSize)
		assert.Equal(t, pbo.Dummy, bank.Files[0].Type)
	})

	t.Run("PBO with multiple files", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// Header for "a.txt"
			'a', '.', 't', 'x', 't', 0, // filename
			0, 0, 0, 0, // type: Dummy
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			3, 0, 0, 0, // data size: 3
			// Header for "b.txt"
			'b', '.', 't', 'x', 't', 0, // filename
			0, 0, 0, 0, // type: Dummy
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			4, 0, 0, 0, // data size: 4
			// End header
			0,          // no filename
			0, 0, 0, 0, // type
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// data for a.txt
			'a', 'b', 'c',
			// data for b.txt
			'd', 'e', 'f', 'g',
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 2)
		assert.Equal(t, "a.txt", bank.Files[0].Filename)
		assert.Equal(t, uint32(3), bank.Files[0].DataSize)
		assert.Equal(t, "b.txt", bank.Files[1].Filename)
		assert.Equal(t, uint32(4), bank.Files[1].DataSize)
	})

	t.Run("PBO with properties and files", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{
			// Vers header
			0,                      // no filename
			0x73, 0x72, 0x65, 0x56, // type: "Vers"
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// properties
			'p', 'r', 'e', 'f', 'i', 'x', 0, // key: "prefix"
			'm', 'o', 'd', 0, // value: "mod"
			0, // terminal empty string
			// File header: "readme.txt"
			'r', 'e', 'a', 'd', 'm', 'e', '.', 't', 'x', 't', 0, // filename
			0, 0, 0, 0, // type: Dummy
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			2, 0, 0, 0, // data size: 2
			// End header
			0,          // no filename
			0, 0, 0, 0, // type
			0, 0, 0, 0, // original size
			0, 0, 0, 0, // offset
			0, 0, 0, 0, // timestamp
			0, 0, 0, 0, // data size
			// data
			'o', 'k',
		})

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, "mod", bank.Properties["prefix"])
		require.Len(t, bank.Files, 1)
		assert.Equal(t, "readme.txt", bank.Files[0].Filename)
	})
}
