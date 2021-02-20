package file_reader

import (
        "io"
	"log_monitor/monitor/core"
	"os"
)

func ReadLastNLinesFromFile(filename string, numLines uint64) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return core.ReadLastNLines(file, numLines)
}

func ReadLastLinesContainsStringFromFile(filename string, expr string) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return core.ReadLastLinesContainsString(file, expr)
}
