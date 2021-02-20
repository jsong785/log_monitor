package core

import (
	"bytes"
	"io"
	"log_monitor/monitor/core_utils"
	"strings"
)

func ReadReverseNLines(buffer io.ReadSeeker, numLines uint64) (io.ReadSeeker, error) {
	if numLines == 0 {
		return nil, nil
	}

	count := uint64(0)
	return readReverse(
		buffer,
		func(string) bool {
			count++
			return true
		},
		func() (bool, error) {
			return count < numLines, nil
		})
}

func ReadReversePassesFilter(buffer io.ReadSeeker, expr string) (io.ReadSeeker, error) {
	return readReverse(
		buffer,
		func(line string) bool {
			return strings.Contains(line, expr)
		},
		func() (bool, error) {
			pos, err := buffer.Seek(0, io.SeekCurrent)
			return pos > 0, err
		})
}

func readLineReverse(buffer io.ReadSeeker) (string, error) {
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

	return string(core_utils.ReverseBytes(reverseBuffer)), nil
}

func readReverse(buffer io.ReadSeeker, isValid func(string) bool, keepReading func() (bool, error)) (io.ReadSeeker, error) {
	var results bytes.Buffer
	for {
		line, err := readLineReverse(buffer)
		if err != nil {
			return nil, err
		} else if isValid(line) {
			results.WriteString(line)
		}

		ok, err := keepReading()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}
	}
	return bytes.NewReader(results.Bytes()), nil
}