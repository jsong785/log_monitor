package reader

import (
        "bytes"
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

func ReadLinesInReverse(buffer io.ReadSeeker, isValid func(string) bool, keepReading func() (bool, error)) (io.ReadSeeker, error) {
        var results bytes.Buffer
	for {
		line, err := ReadLineReverse(buffer)
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

func ReadLastNLinesHelper(buffer io.ReadSeeker, numLines uint64) (io.ReadSeeker, error) {
	if numLines == 0 {
                return nil, nil
	}

        count := uint64(0)
	return ReadLinesInReverse(
		buffer,
		func(string) bool {
                        count++
			return true
		},
		func() (bool, error) {
                        return count < numLines, nil
		})
}

func ReadLastNLines(buffer io.ReadSeeker, numLines uint64) (io.ReadSeeker, error) {
	_, err := buffer.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	return ReadLastNLinesHelper(buffer, numLines)
}

func ReadLastLinesContainsStringHelper(buffer io.ReadSeeker, expr string) (io.ReadSeeker, error) {
	return ReadLinesInReverse(
		buffer,
		func(line string) bool {
			return strings.Contains(line, expr)
		},
		func() (bool, error) {
			pos, err := buffer.Seek(0, io.SeekCurrent)
			return pos > 0, err
		})
}

func ReadLastLinesContainsString(buffer io.ReadSeeker, expr string) (io.ReadSeeker, error) {
	_, err := buffer.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
 return ReadLastLinesContainsStringHelper(buffer, expr)
}
