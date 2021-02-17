package log_monitor

import (
    "io"
    "strings"
)

func ReadLineReverse(buffer io.ReadSeeker) (string, error) {
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
        buffer.Seek(-1, io.SeekCurrent)

        if charBuffer[0] == '\n' {
            break
        }
        reverseBuffer = append(reverseBuffer, charBuffer[0])
        if pos == 0 {
            break
        }
    }

    var builder strings.Builder
    for i := len(reverseBuffer)-1; i >= 0; i-- {
        builder.WriteByte(reverseBuffer[i])
    }
    return builder.String(), nil
}

func ReadLastNLines(buffer io.ReadSeeker, numLines uint64) ([]string, error) {
    _, err := buffer.Seek(0, io.SeekEnd)
    if err != nil {
        return nil, err
    }

    linesRead := make([]string, 0)
    for i := uint64(0); i < numLines; i++ {
        line, err := ReadLineReverse(buffer)
        if err != nil {
            return nil, err
        }
        linesRead = append(linesRead, line)
    }
    return linesRead, nil
}

