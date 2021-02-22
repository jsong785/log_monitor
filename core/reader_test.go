package core

import (
	"errors"
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"log_monitor/monitor/core_utils"
	"log_monitor/monitor/test_utils"
	"strings"
	"testing"
)

func TestreadLineReverseFast(t *testing.T) {
	//empty
	func() {
		reader := strings.NewReader("")
		reader.Seek(0, io.SeekEnd)

		line, err := readLineReverseFast(reader)
		assert.Equal(t, 0, len(line))
		assert.NotNil(t, err)
	}()

	// empty new line
	func() {
		reader := strings.NewReader("\n")
		reader.Seek(0, io.SeekEnd)

		line, err := readLineReverseFast(reader)
		assert.Equal(t, "\n", line)
		assert.Nil(t, err)
	}()

	// 3 lines
	func() {
		reader := strings.NewReader("123\n456\n\n")
		reader.Seek(0, io.SeekEnd)

		line, err := readLineReverseFast(reader)
		assert.Equal(t, "\n", line)
		assert.Nil(t, err)

		line, err = readLineReverseFast(reader)
		assert.Equal(t, "456\n", line)
		assert.Nil(t, err)

		line, err = readLineReverseFast(reader)
		assert.Equal(t, "123\n", line)
		assert.Nil(t, err)

		line, err = readLineReverseFast(reader)
		assert.Equal(t, 0, len(line))
		assert.NotNil(t, err)
	}()
}

func TestReadReverseNLinesFast(t *testing.T) {
	// empty
	func () {
		reader := strings.NewReader("")
		reader.Seek(0, io.SeekEnd)
		
		res, err := ReadReverseNLinesFast(reader, 0)
		assert.Nil(t, res)
		assert.Nil(t, err)

		reader.Seek(0, io.SeekEnd)
		res, err = ReadReverseNLinesFast(reader, 1)
		assert.Nil(t, res)
		assert.NotNil(t, err)

		reader.Seek(0, io.SeekEnd)
		res, err = ReadReverseNLinesFast(reader, 2)
		assert.Nil(t, res)
		assert.NotNil(t, err)
	}()

	// not empty
	func () {
		getString := func(reader io.ReadSeeker) string {
			var buffer bytes.Buffer
			buffer.ReadFrom(reader)
			return buffer.String()
		}

		reader := strings.NewReader("123\n456\n789\n\n")
		reader.Seek(0, io.SeekEnd)
		
		res, err := ReadReverseNLinesFast(reader, 0)
		assert.Nil(t, res)
		assert.Nil(t, err)

		res, err = ReadReverseNLinesFast(reader, 1)
		assert.Equal(t, "\n", getString(res))
		assert.Nil(t, err)

		res, err = ReadReverseNLinesFast(reader, 2)
		assert.Equal(t, "789\n456\n", getString(res))
		assert.Nil(t, err)

		res, err = ReadReverseNLinesFast(reader, 1)
		assert.Equal(t, "123\n", getString(res))
		assert.Nil(t, err)

		res, err = ReadReverseNLinesFast(reader, 1)
		assert.Nil(t,res )
		assert.NotNil(t, err)
	}()
}

func TestReadReversePassesFilter(t *testing.T) {
	// empty
	func () {
		reader := strings.NewReader("")
		reader.Seek(0, io.SeekEnd)
		
		res, err := ReadReversePassesFilter(reader, "")
		assert.Nil(t, res)
		assert.NotNil(t, err)

		reader.Seek(0, io.SeekEnd)
		res, err = ReadReversePassesFilter(reader, "abc")
		assert.Nil(t, res)
		assert.NotNil(t, err)

		reader.Seek(0, io.SeekEnd)
		res, err = ReadReversePassesFilter(reader, "def")
		assert.Nil(t, res)
		assert.NotNil(t, err)
	}()

	// not empty
	func () {
		getString := func(reader io.ReadSeeker) string {
			var buffer bytes.Buffer
			buffer.ReadFrom(reader)
			return buffer.String()
		}

		reader := strings.NewReader("pass\nabc\npassabc\ndef\npassdef\n\n")
		reader.Seek(0, io.SeekEnd)
		
		res, err := ReadReversePassesFilter(reader, "pass")
		assert.Equal(t, "passdef\npassabc\npass\n", getString(res))
		assert.Nil(t, err)

		res, err = ReadReversePassesFilter(reader, "pass")
		assert.Nil(t, res)
		assert.NotNil(t, err)
	}()
}

func TestreadLineReverse_Empty(t *testing.T) {
	reader := strings.NewReader("")
	reader.Seek(0, io.SeekEnd)

	line, err := readLineReverse(reader)
	assert.Equal(t, 0, len(line))
	assert.NotNil(t, err)
}

func TestreadLineReverse_EmptyLines(t *testing.T) {
	reader := strings.NewReader("\n\n")

	_, err := reader.Seek(0, io.SeekEnd)
	assert.Nil(t, err)

	// read first line
	func() {
		line, err := readLineReverse(reader)
		assert.Equal(t, "\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(1), pos)
		assert.Nil(t, err)
	}()
	// read second line
	func() {
		line, err := readLineReverse(reader)
		assert.Equal(t, "\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
		assert.Nil(t, err)
	}()
	// read non-existent third line
	func() {
		line, err := readLineReverse(reader)
		assert.Equal(t, 0, len(line))
		assert.NotNil(t, err)
	}()
}

func TestreadLineReverse_NonEmptyLines(t *testing.T) {
	reader := strings.NewReader("abc\ndef\n")

	_, err := reader.Seek(0, io.SeekEnd)
	assert.Nil(t, err)

	// read first line
	func() {
		line, err := readLineReverse(reader)
		assert.Equal(t, "def\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(4), pos)
		assert.Nil(t, err)
	}()
	// read second line
	func() {
		line, err := readLineReverse(reader)
		assert.Equal(t, "abc\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
		assert.Nil(t, err)
	}()
	// read non-existent third line
	func() {
		line, err := readLineReverse(reader)
		assert.Equal(t, 0, len(line))
		assert.NotNil(t, err)
	}()
}

func TestreadReverse_Stop(t *testing.T) {
	reader := strings.NewReader("abc\ndef\n")

	// read all lines
	reader.Seek(0, io.SeekEnd)
	func() {
		count := 0
		lines, err := readReverse(
			readLineReverse,
			reader,
			func(string) bool {
				count++
				return true
			},
			func() (bool, error) {
				return count < 2, nil
			})
		assert.Equal(t, []string{"def\n", "abc\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

	// read one line
	reader.Seek(0, io.SeekEnd)
	func() {
		count := 0
		lines, err := readReverse(
			readLineReverse,
			reader,
			func(string) bool {
				count++
				return true
			},
			func() (bool, error) {
				return count < 1, nil
			})
		assert.Equal(t, []string{"def\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

	// skip every other
	reader.Reset("abc\ndef\nghi\njkl\n")
	reader.Seek(0, io.SeekEnd)
	func() {
		count := 0
		valid := true
		lines, err := readReverse(
			readLineReverse,
			reader,
			func(string) bool {
				v := valid
				valid = !valid
				return v
			},
			func() (bool, error) {
				count++
				return count < 4, nil
			})
		assert.Equal(t, []string{"jkl\n", "def\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

	// if it starts with an 'a', include it
	reader.Reset("apple\ncar\ndefer\nairplane\nzebra\natom\n")
	reader.Seek(0, io.SeekEnd)
	func() {
		lines, err := readReverse(
			readLineReverse,
			reader,
			func(val string) bool {
				return len(val) > 0 && val[0] == 'a'
			},
			func() (bool, error) {
				pos, err := reader.Seek(0, io.SeekCurrent)
				return pos > 0, err
			})
		assert.Equal(t, []string{"atom\n", "airplane\n", "apple\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

	// if it never stops, it will error
	reader.Seek(0, io.SeekEnd)
	func() {
		lines, err := readReverse(
			readLineReverse,
			reader,
			func(string) bool {
				return true
			},
			func() (bool, error) {
				return true, nil
			})
		assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()

	// if it errors, just return blank, even if the return is true
	reader.Seek(0, io.SeekEnd)
	func() {
		count := 0
		lines, err := readReverse(
			readLineReverse,
			reader,
			func(string) bool {
				return true
			},
			func() (bool, error) {
				count++
				if count < 2 {
					return true, nil
				} else {
					return true, errors.New("some_error_in_a_test")
				}
			})
		assert.Nil(t, lines)
		assert.NotNil(t, err)
		assert.Equal(t, "some_error_in_a_test", err.Error())
	}()
}

func TestReadReverseNLines_Empty(t *testing.T) {
	reader := strings.NewReader("")

	func() {
		lines, err := ReadReverseNLines(reader, 0)
		assert.Nil(t, lines)
		assert.Nil(t, err)
	}()

	func() {
		lines, err := ReadReverseNLines(reader, 2)
		assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

func TestReadReverseNLines_NotEmpty(t *testing.T) {
	reader := strings.NewReader("abc\ndef\nghi\njkl\n")

	_, err := core_utils.SeekEnd(reader)
	assert.Nil(t, err)
	func() {
		lines, err := ReadReverseNLines(reader, 0)
		assert.Nil(t, lines)
		assert.Nil(t, err)
	}()

	_, err = core_utils.SeekEnd(reader)
	assert.Nil(t, err)
	func() {
		lines, err := ReadReverseNLines(reader, 1)
		assert.Equal(t, []string{"jkl\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

	_, err = core_utils.SeekEnd(reader)
	assert.Nil(t, err)
	func() {
		lines, err := ReadReverseNLines(reader, 2)
		assert.Equal(t, []string{"jkl\n", "ghi\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()
}

func TestReadReversePassesFilter_Empty(t *testing.T) {
	reader := strings.NewReader("")

	func() {
		lines, err := ReadReversePassesFilter(reader, "")
		assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()

	func() {
		lines, err := ReadReversePassesFilter(reader, "abc")
		assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

func TestReadReversePassesFilter_NotEmpty(t *testing.T) {
	reader := strings.NewReader("one\ntwo\nthree\nfour\n")

	_, err := core_utils.SeekEnd(reader)
	assert.Nil(t, err)
	func() {
		lines, err := ReadReversePassesFilter(reader, "")
		assert.Equal(t, []string{"four\n", "three\n", "two\n", "one\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

	_, err = core_utils.SeekEnd(reader)
	assert.Nil(t, err)
	func() {
		lines, err := ReadReversePassesFilter(reader, "o")
		assert.Equal(t, []string{"four\n", "two\n", "one\n"}, test_utils.GetLines(lines))
		assert.Nil(t, err)
	}()

}
