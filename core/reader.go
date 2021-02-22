package core

import (
	"bytes"
	"io"
	"log_monitor/monitor/core_utils"
	"strings"
)

func ReadReverseNLinesFast(buffer io.ReadSeeker, numLines uint64) (io.ReadSeeker, error) {
	return readReverseNLinesHelper(buffer, numLines, true)
}

func ReadReverseNLines(buffer io.ReadSeeker, numLines uint64) (io.ReadSeeker, error) {
	return readReverseNLinesHelper(buffer, numLines, false)
}

func ReadReversePassesFilter(buffer io.ReadSeeker, expr string) (io.ReadSeeker, error) {
	return readReversePassesFilterHelper(buffer, expr,false)
}

func ReadReversePassesFilterFast(buffer io.ReadSeeker, expr string) (io.ReadSeeker, error) {
	return readReversePassesFilterHelper(buffer, expr, true)
}

// sanitary flag true expects perfect new lines;
// does not handle any concurrent changes to file (truncation)
func readReversePassesFilterHelper(buffer io.ReadSeeker, expr string, sanitary bool) (io.ReadSeeker, error) {
	validFunc := func(line string) bool {
		return strings.Contains(line, expr)
	}

	keepReadingFunc := func() (bool, error) {
			pos, err := buffer.Seek(0, io.SeekCurrent)
			return pos > 0, err
	}

	if sanitary {
		return readReverse(readLineReverseFast, buffer, validFunc, keepReadingFunc)
	} else {
		return readReverse(readLineReverse, buffer, validFunc, keepReadingFunc)
	}
}

// sanitary flag true expects perfect new lines;
// does not handle any concurrent changes to file (truncation)
func readReverseNLinesHelper(buffer io.ReadSeeker, numLines uint64, sanitary bool) (io.ReadSeeker, error) {
	if numLines == 0 {
		return nil, nil
	}

	count := uint64(0)
	validFunc := func (string) bool {
				count++
				return true
	}
	keepReadingFunc := func() (bool, error) {
				return count < numLines, nil
	}

	if sanitary {
		return readReverse(readLineReverseFast, buffer, validFunc, keepReadingFunc)
	} else {
		return readReverse(readLineReverse, buffer, validFunc, keepReadingFunc)
	}
}

type reverseLineReader func(io.ReadSeeker) (string, error)
func readReverse(reader reverseLineReader, buffer io.ReadSeeker, isValid func(string) bool, keepReading func() (bool, error)) (io.ReadSeeker, error) {
	var results bytes.Buffer
	for {
		line, err := reader(buffer)
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

// this assumes buffer doesn't change and has perfect lines
func readLineReverseFast(buffer io.ReadSeeker) (string, error) {
	var reverseBuffer []byte
	write := func(start int64, end int64) {
		reverseBuffer = make([]byte, end-start)
		buffer.Read(reverseBuffer)
		buffer.Seek(start-end, io.SeekCurrent)
	}

	end, _ := buffer.Seek(0, io.SeekCurrent)
	foundNewLine := false
	for {
		pos, err := buffer.Seek(-1, io.SeekCurrent)
		if err != nil {
			return "", err
		} else if pos == 0 {
			write(int64(0), end)
			break
		}

		var charBuffer [1]byte
		buffer.Read(charBuffer[:])
		if charBuffer[0] == '\n' {
			if !foundNewLine {
				foundNewLine = true
			} else {
				start, _ := buffer.Seek(0, io.SeekCurrent)
				write(start, end)
				break
			}
		}
		buffer.Seek(-1, io.SeekCurrent)
	}
	return string(reverseBuffer), nil
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