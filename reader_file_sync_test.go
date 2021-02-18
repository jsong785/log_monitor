package log_monitor

import (
	"errors"
	"github.com/stretchr/testify/assert"
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
		lines, err := ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, lines)
	}()

	// remove the file, it doesn't exist, but fd keeps it alive
	func() {
		assert.Nil(t, os.Remove(filename))
		assert.False(t, DoesFileExist(filename))

		lines, err := ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ghi\n", "def\n"}, lines)
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
		lines, err := ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"five\n"}, lines)
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

		lines, err := ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"four\n", "three\n"}, lines)
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
		lines, err := ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"five\n"}, lines)
	}()

	// write
	func() {
		writer.WriteString("six\nseven\neight\nnine\nten\n")
		writer.Sync() // commit changes

		lines, err := ReadLastNLinesHelper(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"four\n", "three\n"}, lines)
	}()

	func() {
		lines, err := ReadLastNLines(file, 9)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ten\n", "nine\n", "eight\n", "seven\n", "six\n", "five\n", "four\n", "three\n", "two\n"}, lines)
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
		lines, err := ReadLastNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, lines)
	}()

	// truncate the file
	func() {
		assert.Nil(t, os.Truncate(filename, 0))

		lines, err := ReadLastNLinesHelper(file, 2)
		assert.NotNil(t, err)
		assert.Equal(t, 0, len(lines))
	}()
}
