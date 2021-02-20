package test_utils

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestGetString(t *testing.T) {
	reader := strings.NewReader("12345")
	cache, err := reader.Seek(3, io.SeekStart)
	assert.Nil(t, err)

	res := GetString(reader)
	assert.Equal(t, "12345", res)

	cur, err := reader.Seek(0, io.SeekCurrent)
	assert.Nil(t, err)
	assert.Equal(t, cache, cur)
}

func TestGetLines(t *testing.T) {
	reader := strings.NewReader("abc\ndef\nghi\n")
	cache, err := reader.Seek(3, io.SeekStart)
	assert.Nil(t, err)

	res := GetLines(reader)
	assert.Equal(t, []string{"abc\n", "def\n", "ghi\n"}, res)

	cur, err := reader.Seek(0, io.SeekCurrent)
	assert.Nil(t, err)
	assert.Equal(t, cache, cur)
}

func TestGetLen(t *testing.T) {
	reader := strings.NewReader("abcdef")
	cache, err := reader.Seek(3, io.SeekStart)
	assert.Nil(t, err)

	res := GetLen(reader)
	assert.Equal(t, 6, res)

	cur, err := reader.Seek(0, io.SeekCurrent)
	assert.Nil(t, err)
	assert.Equal(t, cache, cur)
}

func TestResetSeeker(t *testing.T) {
	reader := strings.NewReader("abcdef")

	reset := resetSeeker(reader)

	_, err := reader.Seek(3, io.SeekStart)
	assert.Nil(t, err)

	reset()
	cur, err := reader.Seek(0, io.SeekStart)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), cur)
}
