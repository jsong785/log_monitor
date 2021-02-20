package file_reader

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log_monitor/monitor/core"
	"log_monitor/monitor/core_utils"
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
	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)
	func() {
		lines, err := core.ReadReverseNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// remove the file, it doesn't exist, but fd keeps it alive
	func() {
		assert.Nil(t, os.Remove(filename))
		assert.False(t, DoesFileExist(filename))

		lines, err := core.ReadReverseNLines(file, 2)
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
	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadReverseNLines(file, 1)
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

		lines, err := core.ReadReverseNLines(file, 2)
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

	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)
	func() {
		lines, err := core.ReadReverseNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"five\n"}, test_utils.GetLines(lines))
	}()

	// write
	func() {
		writer.WriteString("six\nseven\neight\nnine\nten\n")
		writer.Sync() // commit changes

		lines, err := core.ReadReverseNLines(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"four\n", "three\n"}, test_utils.GetLines(lines))
	}()

	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)
	func() {
		lines, err := core.ReadReverseNLines(file, 9)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ten\n", "nine\n", "eight\n", "seven\n", "six\n", "five\n", "four\n", "three\n", "two\n"}, test_utils.GetLines(lines))
	}()
}

// truncation happens during read; nothing has been written yet so the current position is invalid; fails
func TestTruncateWhileReading_1(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)
	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadReverseNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		assert.Nil(t, os.Truncate(filename, 0))

		lines, err := core.ReadReverseNLines(file, 2)
		assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

// truncation happens during read; things are written but not to the current position; fails
func TestTruncateWhileReading_2(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)
	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadReverseNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		writer, err := os.Create(filename)
		assert.Nil(t, err)
		writer.WriteString("aaa\n")
		writer.Close()

		lines, err := core.ReadReverseNLines(file, 2)
		assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

// file gets truncated and appended to mid-read past the position; current position happens to be \n, things continue as "normal"
func TestTruncateWhileReading_3(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)
	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadReverseNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		writer, err := os.Create(filename)
		assert.Nil(t, err)
		writer.WriteString("aaa\nbbb\nccc\nddd\neee\nfff\nggg\n")
		writer.Close()

		lines, err := core.ReadReverseNLines(file, 2)
		assert.Nil(t, err)
		assert.Less(t, 0, test_utils.GetLen(lines))
		res := test_utils.GetLines(lines)
		assert.Equal(t, []string{"ccc\n", "bbb\n"}, res)
	}()
}

// file gets truncated and appended to mid-read past the position, the next seek is to the next newline
func TestTruncateWhileReading_4(t *testing.T) {
	filename := "test_truncate"
	assert.False(t, DoesFileExist(filename))
	defer os.Remove(filename)
	CreateAndWriteFile(filename, "abc\ndef\nghi\njkl\n")
	assert.True(t, DoesFileExist(filename))

	file, err := os.Open(filename)
	assert.Nil(t, err)
	_, err = core_utils.SeekEnd(file)
	assert.Nil(t, err)

	func() {
		lines, err := core.ReadReverseNLines(file, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"jkl\n", "ghi\n"}, test_utils.GetLines(lines))
	}()

	// truncate the file
	func() {
		writer, err := os.Create(filename)
		assert.Nil(t, err)
		//                 "abc\ndef\n"
		writer.WriteString("th\ne quick brown fox jumps over some fence\nthe quick brown fox jumps over some fence\n")
		writer.Close()

		lines, err := core.ReadReverseNLines(file, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"th\n"}, test_utils.GetLines(lines))
	}()
}
