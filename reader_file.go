package log_monitor

import (
    "os"
)

func ReadLastNLinesFromFile(filename string, numLines uint64) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    return ReadLastNLines(file, numLines)
}

func ReadLastLinesContainsStringFromFile(filename string, expr string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    return ReadLastLinesContainsString(file, expr)
}

