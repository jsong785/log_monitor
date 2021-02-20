package file_reader

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log_monitor/monitor/core"
        "log_monitor/monitor/test_utils"
	"os"
	"testing"
)

func DoesFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func CreateAndWriteFile(filename string, contents string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func TestDeleteWhileReading(t *testing.T) {
	filename := "test_delete"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// remove the file, it doesn't exist, but fd keeps it alive
	func() {
		assert.Nil(t, os.Remove(filename))
		assert.False(t, DoesFileExist(filename))

		lines, err := core.ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ghi\n", "def\n"}, test_utils.GetLines(lines))
	}()
	// it really doesn't exist
	assert.False(t, DoesFileExist(filename))
}

func TestMoveWhileReading(t *testing.T) {
	filename := "test_move"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "one\ntwo\nthree\nfour\nfive\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"five\n"}, test_utils.GetLines(lines))
	}()

	// move the file, fd follows
	defer os.Remove("test_move_candidate")
	func() {
		assert.False(t, DoesFileExist("test_move_candidate"))
		assert.Nil(t, os.Rename(filename, "test_move_candidate"))

		// renamed, and it exists
		assert.True(t, DoesFileExist("test_move_candidate"))
		// doesn't exist
		assert.False(t, DoesFileExist(filename))

		lines, err := core.ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"four\n", "three\n"}, test_utils.GetLines(lines))
	}()
	// it really doesn't exist
	assert.False(t, DoesFileExist(filename))
}

func TestWriteWhileReading(t *testing.T) {
	filename := "test_write"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "one\ntwo\nthree\nfour\nfive\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)

	writer, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	assert.Nil(t, err)
	defer writer.Close()

	func() {
		lines, err := core.ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"five\n"}, test_utils.GetLines(lines))
	}()

	// write
	func() {
		writer.WriteString("six\nseven\neight\nnine\nten\n")
		writer.Sync() // commit changes

		lines, err := core.ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"four\n", "three\n"}, test_utils.GetLines(lines))
	}()

	func() {
		lines, err := core.ReadLastNLines(file, 9)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ten\n", "nine\n", "eight\n", "seven\n", "six\n", "five\n", "four\n", "three\n", "two\n"}, test_utils.GetLines(lines))
	}()
}

func TestTruncateWhileReading(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		assert.Nil(t, os.Truncate(filename, 0))

		lines, err := core.ReadLastNLinesHelper(file, 2)
                assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

func TestTruncateWhileReadingEdgeCase(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		writer, err := os.Create(filename)
		assert.Nil(t, err)
		writer.WriteString("one\ntwo\nthree\nfour\nfive\nsix\nseven\n")
		writer.Close()

		lines, err := core.ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Less(t, 0, test_utils.GetLen(lines))
                _ = lines
	}()
}

func TestTruncateWhileReadingEdgeCase2(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadLastNLines(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n", "ghi\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		writer, err := os.Create(filename)
		assert.Nil(t, err)
		writer.WriteString("aaa\nbbb\n")
		writer.Close()

		lines, err := core.ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"bbb\n", "aaa\n"}, test_utils.GetLines(lines))
	}()
}
