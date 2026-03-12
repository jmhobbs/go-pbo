package pbo_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmhobbs/go-pbo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewBuilder(t *testing.T) {
	b := pbo.NewBuilder()
	assert.NotNil(t, b)
}

func Test_Builder_SetProperty(t *testing.T) {
	t.Run("returns builder for chaining", func(t *testing.T) {
		b := pbo.NewBuilder()
		result := b.SetProperty("key", "value")
		assert.Equal(t, b, result)
	})

	t.Run("multiple properties appear in built PBO", func(t *testing.T) {
		var buf bytes.Buffer
		err := pbo.NewBuilder().
			SetProperty("prefix", "mymod").
			SetProperty("version", "1.0").
			Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, "mymod", bank.Properties["prefix"])
		assert.Equal(t, "1.0", bank.Properties["version"])
	})

	t.Run("overwriting a property keeps last value", func(t *testing.T) {
		var buf bytes.Buffer
		err := pbo.NewBuilder().
			SetProperty("key", "first").
			SetProperty("key", "second").
			Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, "second", bank.Properties["key"])
	})
}

func Test_Builder_AddFile(t *testing.T) {
	t.Run("returns builder for chaining", func(t *testing.T) {
		b := pbo.NewBuilder()
		result := b.AddFile(strings.NewReader("data"), "file.txt", time.Now(), 4)
		assert.Equal(t, b, result)
	})

	t.Run("file appears in built PBO", func(t *testing.T) {
		var buf bytes.Buffer
		err := pbo.NewBuilder().
			AddFile(strings.NewReader("hello"), "greet.txt", time.Unix(1000, 0), 5).
			Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 1)
		assert.Equal(t, "greet.txt", bank.Files[0].Filename)
		assert.Equal(t, uint32(5), bank.Files[0].DataSize)
		assert.Equal(t, uint32(1000), bank.Files[0].Timestamp)
	})

	t.Run("files are stored in PBO output", func(t *testing.T) {
		var buf bytes.Buffer
		err := pbo.NewBuilder().
			AddFile(strings.NewReader("abc"), "a.txt", time.Unix(100, 0), 3).
			AddFile(strings.NewReader("defg"), "b.txt", time.Unix(200, 0), 4).
			Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 2)

		assert.Equal(t, "a.txt", bank.Files[0].Filename)
		assert.Equal(t, uint32(3), bank.Files[0].DataSize)
		assert.Equal(t, uint32(100), bank.Files[0].Timestamp)

		r0, err := bank.Files[0].Reader()
		require.NoError(t, err)
		c0, err := io.ReadAll(r0)
		require.NoError(t, err)
		assert.Equal(t, "abc", string(c0))

		assert.Equal(t, "b.txt", bank.Files[1].Filename)
		assert.Equal(t, uint32(4), bank.Files[1].DataSize)
		assert.Equal(t, uint32(200), bank.Files[1].Timestamp)

		r1, err := bank.Files[1].Reader()
		require.NoError(t, err)
		c1, err := io.ReadAll(r1)
		require.NoError(t, err)
		assert.Equal(t, "defg", string(c1))
	})
}

func Test_Builder_AddFileFromPath(t *testing.T) {
	t.Run("returns builder for chaining", func(t *testing.T) {
		b := pbo.NewBuilder()
		result := b.AddFileFromPath("somefile.txt")
		assert.Equal(t, b, result)
	})

	t.Run("file appears in built PBO with correct content", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "data.txt")
		require.NoError(t, os.WriteFile(path, []byte("from disk"), 0600))

		var buf bytes.Buffer
		err := pbo.NewBuilder().
			AddFileFromPath(path).
			Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.Len(t, bank.Files, 1)
		assert.Equal(t, uint32(9), bank.Files[0].DataSize)

		r, err := bank.Files[0].Reader()
		require.NoError(t, err)
		content, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, "from disk", string(content))
	})

	t.Run("non-existent file causes Build to error", func(t *testing.T) {
		var buf bytes.Buffer
		err := pbo.NewBuilder().
			AddFileFromPath("/nonexistent/path/file.txt").
			Build(&buf)
		require.Error(t, err)
	})

	t.Run("directory path causes Build to error", func(t *testing.T) {
		tmp := t.TempDir()

		var buf bytes.Buffer
		err := pbo.NewBuilder().
			AddFileFromPath(tmp).
			Build(&buf)
		require.Error(t, err)
	})
}

func Test_Builder_Build(t *testing.T) {
	t.Run("minimal PBO with no properties or files is loadable", func(t *testing.T) {
		var buf bytes.Buffer
		err := pbo.NewBuilder().Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		assert.Empty(t, bank.Files)
	})

	t.Run("round-trip preserves properties and file content", func(t *testing.T) {
		tmp := t.TempDir()
		diskFile := filepath.Join(tmp, "disk.txt")
		require.NoError(t, os.WriteFile(diskFile, []byte("on disk"), 0600))

		var buf bytes.Buffer
		err := pbo.NewBuilder().
			SetProperty("prefix", "testmod").
			AddFile(strings.NewReader("in memory"), "mem.txt", time.Unix(42, 0), 9).
			AddFileFromPath(diskFile).
			Build(&buf)
		require.NoError(t, err)

		bank, err := pbo.Load(bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)

		assert.Equal(t, "testmod", bank.Properties["prefix"])
		require.Len(t, bank.Files, 2)

		// AddFileFromPath entries are written before AddFile entries
		r0, err := bank.Files[0].Reader()
		require.NoError(t, err)
		c0, err := io.ReadAll(r0)
		require.NoError(t, err)
		assert.Equal(t, "on disk", string(c0))

		r1, err := bank.Files[1].Reader()
		require.NoError(t, err)
		c1, err := io.ReadAll(r1)
		require.NoError(t, err)
		assert.Equal(t, "in memory", string(c1))
	})
}
