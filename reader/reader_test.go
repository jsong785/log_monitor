package reader

import (
        "bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func SplitReader(reader io.ReadSeeker) []string {
    var buffer bytes.Buffer
    buffer.ReadFrom(reader)
    list := strings.SplitAfter(buffer.String(), "\n")
    if len(list) > 0 {
        list = list[:len(list)-1]
    }
    return list
}

func TestReadLineReverse_Empty(t *testing.T) {
	reader := strings.NewReader("")
	reader.Seek(0, io.SeekEnd)

	line, err := ReadLineReverse(reader)
        assert.Equal(t, 0, len(line))
	assert.NotNil(t, err)
}

func TestReadLineReverse_EmptyLines(t *testing.T) {
	reader := strings.NewReader("\n\n")

	_, err := reader.Seek(0, io.SeekEnd)
	assert.Nil(t, err)

	// read first line
	func() {
		line, err := ReadLineReverse(reader)
		assert.Equal(t, "\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(1), pos)
		assert.Nil(t, err)
	}()
	// read second line
	func() {
		line, err := ReadLineReverse(reader)
		assert.Equal(t, "\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
		assert.Nil(t, err)
	}()
	// read non-existent third line
	func() {
		line, err := ReadLineReverse(reader)
                assert.Equal(t, 0, len(line))
		assert.NotNil(t, err)
	}()
}

func TestReadLineReverse_NonEmptyLines(t *testing.T) {
	reader := strings.NewReader("abc\ndef\n")

	_, err := reader.Seek(0, io.SeekEnd)
	assert.Nil(t, err)

	// read first line
	func() {
		line, err := ReadLineReverse(reader)
		assert.Equal(t, "def\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(4), pos)
		assert.Nil(t, err)
	}()
	// read second line
	func() {
		line, err := ReadLineReverse(reader)
		assert.Equal(t, "abc\n", line)
		assert.Nil(t, err)

		pos, err := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
		assert.Nil(t, err)
	}()
	// read non-existent third line
	func() {
		line, err := ReadLineReverse(reader)
                assert.Equal(t, 0, len(line))
		assert.NotNil(t, err)
	}()
}

func TestReadLinesInReverse_Stop(t *testing.T) {
	reader := strings.NewReader("abc\ndef\n")

	// read all lines
	reader.Seek(0, io.SeekEnd)
	func() {
                count := 0
		lines, err := ReadLinesInReverse(reader,
			func(string) bool {
                                count++
				return true
			},
			func() (bool, error) {
                                return count < 2 , nil
			})
		assert.Equal(t, []string{"def\n", "abc\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

	// read one line
	reader.Seek(0, io.SeekEnd)
	func() {
                count := 0
		lines, err := ReadLinesInReverse(reader,
			func(string) bool {
                                count++
				return true
			},
			func() (bool, error) {
                                return count < 1, nil
			})
		assert.Equal(t, []string{"def\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

	// skip every other
	reader.Reset("abc\ndef\nghi\njkl\n")
	reader.Seek(0, io.SeekEnd)
	func() {
		count := 0
		valid := true
		lines, err := ReadLinesInReverse(reader,
			func(string) bool {
				v := valid
				valid = !valid
				return v
			},
			func() (bool, error) {
				count++
				return count < 4, nil
			})
		assert.Equal(t, []string{"jkl\n", "def\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

	// if it starts with an 'a', include it
	reader.Reset("apple\ncar\ndefer\nairplane\nzebra\natom\n")
	reader.Seek(0, io.SeekEnd)
	func() {
		lines, err := ReadLinesInReverse(reader,
			func(val string) bool {
				return len(val) > 0 && val[0] == 'a'
			},
			func() (bool, error) {
				pos, err := reader.Seek(0, io.SeekCurrent)
				return pos > 0, err
			})
		assert.Equal(t, []string{"atom\n", "airplane\n", "apple\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

	// if it never stops, it will error
	reader.Seek(0, io.SeekEnd)
	func() {
		lines, err := ReadLinesInReverse(reader,
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
		lines, err := ReadLinesInReverse(reader,
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

func TestReadLastNLines_Empty(t *testing.T) {
	reader := strings.NewReader("")

	func() {
		lines, err := ReadLastNLines(reader, 0)
                assert.Nil(t, lines)
		assert.Nil(t, err)
	}()

	func() {
		lines, err := ReadLastNLines(reader, 2)
                assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

func TestReadLastNLines_NotEmpty(t *testing.T) {
	reader := strings.NewReader("abc\ndef\nghi\njkl\n")

	func() {
		lines, err := ReadLastNLines(reader, 0)
                assert.Nil(t, lines)
		assert.Nil(t, err)
	}()

	func() {
		lines, err := ReadLastNLines(reader, 1)
		assert.Equal(t, []string{"jkl\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

	func() {
		lines, err := ReadLastNLines(reader, 2)
		assert.Equal(t, []string{"jkl\n", "ghi\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()
}

func TestReadLastLinesContainsString_Empty(t *testing.T) {
	reader := strings.NewReader("")

	func() {
		lines, err := ReadLastLinesContainsString(reader, "")
                assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()

	func() {
		lines, err := ReadLastLinesContainsString(reader, "abc")
                assert.Nil(t, lines)
		assert.NotNil(t, err)
	}()
}

func TestReadLastLinesContainsString_NotEmpty(t *testing.T) {
	reader := strings.NewReader("one\ntwo\nthree\nfour\n")

	func() {
		lines, err := ReadLastLinesContainsString(reader, "")
		assert.Equal(t, []string{"four\n", "three\n", "two\n", "one\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

	func() {
		lines, err := ReadLastLinesContainsString(reader, "o")
		assert.Equal(t, []string{"four\n", "two\n", "one\n"}, SplitReader(lines))
		assert.Nil(t, err)
	}()

}
