package log_monitor

import (
	"io"
	"strings"
)

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

	var builder strings.Builder
	for i := len(reverseBuffer) - 1; i >= 0; i-- {
		builder.WriteByte(reverseBuffer[i])
	}
	return builder.String(), nil
}

func ReadLinesInReverse(buffer io.ReadSeeker, isValid func(string) bool, keepReading func([]string)bool) ([]string, error) {
    var lines []string
    for {
        line, err := ReadLineReverse(buffer)
        if err != nil {
            return nil, err
        } else if isValid(line) {
            lines = append(lines, line)
        }

        if !keepReading(lines) {
            break;
        }
    }
    return lines, nil
}

func ReadLastNLines(buffer io.ReadSeeker, numLines uint64) ([]string, error) {
    _, err := buffer.Seek(0, io.SeekEnd)
    if err != nil {
            return nil, err
    }
    if numLines == 0 {
        return nil, nil
    }

    return ReadLinesInReverse(
        buffer,
        func(string) bool { 
            return true 
        },
        func(lines []string) bool {
            return uint64(len(lines)) < numLines
        })
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
        func(lines []string) bool {
            pos, err := buffer.Seek(0, io.SeekCurrent)
            return err == nil && pos > 0
        })
}

