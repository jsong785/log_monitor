package reader

import (
	"io"
	"strings"
)

func ReverseSlice(slice []byte) []byte {
	if len(slice) == 0 {
		return slice
	}
	start := 0
	end := len(slice) - 1
	for start < end {
		a := slice[start]
		b := slice[end]

		slice[start] = b
		slice[end] = a
		start++
		end--
	}
	return slice
}

func ReadLineReverse(buffer io.ReadSeeker) (string, error) {
	newlineFound := false

	var reverseBuffer []byte
	for {
		pos, err := buffer.Seek(-1, io.SeekCurrent)
		if err != nil {
			return "", err
		}

		var charBuffer [1]byte
		if _, err := buffer.Read(charBuffer[:]); err != nil {
			return "", err
		}

		if charBuffer[0] == '\n' {
			if newlineFound {
				break
			} else {
				newlineFound = true
			}
		}
		buffer.Seek(-1, io.SeekCurrent)

		reverseBuffer = append(reverseBuffer, charBuffer[0])
		if pos == 0 {
			break
		}
	}

	return string(ReverseSlice(reverseBuffer)), nil
}

func ReadLinesInReverse(buffer io.ReadSeeker, isValid func(string) bool, keepReading func([]string) (bool, error)) ([]string, error) {
	var lines []string
	for {
		line, err := ReadLineReverse(buffer)
		if err != nil {
			return nil, err
		} else if isValid(line) {
			lines = append(lines, line)
		}

		ok, err := keepReading(lines)
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
	}
	return lines, nil
}

func ReadLastNLinesHelper(buffer io.ReadSeeker, numLines uint64) ([]string, error) {
	if numLines == 0 {
		return nil, nil
	}

	return ReadLinesInReverse(
		buffer,
		func(string) bool {
			return true
		},
		func(lines []string) (bool, error) {
			return uint64(len(lines)) < numLines, nil
		})
}

func ReadLastNLines(buffer io.ReadSeeker, numLines uint64) ([]string, error) {
	_, err := buffer.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	return ReadLastNLinesHelper(buffer, numLines)
}

func ReadLastLinesContainsString(buffer io.ReadSeeker, expr string) ([]string, error) {
	_, err := buffer.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	return ReadLinesInReverse(
		buffer,
		func(line string) bool {
			return strings.Contains(line, expr)
		},
		func(lines []string) (bool, error) {
			pos, err := buffer.Seek(0, io.SeekCurrent)
			return pos > 0, err
		})
}
