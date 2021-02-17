package log_monitor

import (
    "io"
    "strings"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestReadLineReverse_Empty(t *testing.T) {
    reader := strings.NewReader("")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, 0, len(line))
    assert.NotNil(t, err)
}

func TestReadLineReverse_EmptyLine(t *testing.T) {
    reader := strings.NewReader("\n")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, 0, len(line))
    assert.Nil(t, err)
}

func TestReadLineReverse_EmptyLine_2(t *testing.T) {
    reader := strings.NewReader("\n\n")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, 0, len(line))
    assert.Nil(t, err)
}

func TestReadLineReverse_EmptyLine_3(t *testing.T) {
    reader := strings.NewReader("123\n")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, 0, len(line))
    assert.Nil(t, err)
}

func TestReadLineReverse_EmptyLine_4(t *testing.T) {
    reader := strings.NewReader("abc\n")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, 0, len(line))
    assert.Nil(t, err)
}

func TestReadLineReverse_Valid_incomplete_line(t *testing.T) {
    reader := strings.NewReader("abc")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, line, "abc")
    assert.Nil(t, err)
}

func TestReadLineReverse_Valid(t *testing.T) {
    reader := strings.NewReader("\nabc")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, line, "abc")
    assert.Nil(t, err)
}

func TestReadLineReverse_Valid_multiple(t *testing.T) {
    reader := strings.NewReader("abc\ndef\nghi")
    reader.Seek(0, io.SeekEnd)

    line, err := ReadLineReverse(reader)
    assert.Equal(t, "ghi", line)
    assert.Nil(t, err)

    line, err = ReadLineReverse(reader)
    assert.Equal(t, "def", line)
    assert.Nil(t, err)

    line, err = ReadLineReverse(reader)
    assert.Equal(t, "abc", line)
    assert.Nil(t, err)
}

func TestReadLastNLines_Empty(t *testing.T) {
    reader := strings.NewReader("")
    line, err := ReadLastNLines(reader, 0)
    assert.Equal(t, 0, len(line))
    assert.Nil(t, err)
}

func TestReadLastNLines_Empty_2(t *testing.T) {
    reader := strings.NewReader("abcd")
    line, err := ReadLastNLines(reader, 0)
    assert.Equal(t, 0, len(line))
    assert.Nil(t, err)
}

func TestReadLastNLines_valid(t *testing.T) {
    reader := strings.NewReader("abc")
    line, err := ReadLastNLines(reader, 1)
    assert.Equal(t, []string{ "abc"}, line)
    assert.Nil(t, err)
}

func TestReadLastNLines_valid_2(t *testing.T) {
    reader := strings.NewReader("abc\ndef\nghi")
    line, err := ReadLastNLines(reader, 3)
    assert.Equal(t, []string{ "ghi", "def", "abc" }, line)
    assert.Nil(t, err)
}

func TestReadLastNLines_valid_3(t *testing.T) {
    reader := strings.NewReader("abc\ndef\nghi")
    line, err := ReadLastNLines(reader, 2)
    assert.Equal(t, []string{ "ghi", "def" }, line)
    assert.Nil(t, err)
}

func TestReadLastNLines_too_many(t *testing.T) {
    reader := strings.NewReader("abc\ndef\nghi")
    line, err := ReadLastNLines(reader, 4)
    assert.Equal(t, 0, len(line))
    assert.NotNil(t, err)
}

