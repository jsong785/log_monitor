package test_utils

import (
	"bytes"
	"io"
	"strings"
)

func GetString(reader io.ReadSeeker) string {
	reset := resetSeeker(reader)
	defer reset()

	reader.Seek(0, io.SeekStart)
	var builder strings.Builder
	io.Copy(&builder, reader)
	return builder.String()
}

func GetLines(reader io.ReadSeeker) []string {
	reset := resetSeeker(reader)
	defer reset()

	reader.Seek(0, io.SeekStart)
	var buffer bytes.Buffer
	buffer.ReadFrom(reader)

	list := strings.SplitAfter(buffer.String(), "\n")
	if len(list) > 0 {
		list = list[:len(list)-1]
	}
	return list
}

func GetLen(reader io.ReadSeeker) int {
	reset := resetSeeker(reader)
	defer reset()

	sz, _ := reader.Seek(0, io.SeekEnd)
	return int(sz)
}

func resetSeeker(reader io.Seeker) func() {
	cache, _ := reader.Seek(0, io.SeekCurrent)
	return func() {
		reader.Seek(cache, io.SeekStart)
	}
}
