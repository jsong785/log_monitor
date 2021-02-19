package reader_file

import (
	"log_monitor/monitor/reader"
	"os"
)

func ReadLastNLinesFromFile(filename string, numLines uint64) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return reader.ReadLastNLines(file, numLines)
}

func ReadLastLinesContainsStringFromFile(filename string, expr string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return reader.ReadLastLinesContainsString(file, expr)
}
