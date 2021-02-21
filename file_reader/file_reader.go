package file_reader

import (
	"io"
	"log_monitor/monitor/core"
	"log_monitor/monitor/core_utils"
	"os"
)

func PocReadReverseNLines(filename string, numLines uint64) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := core_utils.SeekEnd(file)
	return core_utils.LogFuncBind(buffer, err, func(b io.ReadSeeker) (io.ReadSeeker, error) {
		return core.PocReverseNLines(b, numLines)
	})
}

func PocReadReverseNLinesNew(filename string, numLines uint64) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := core_utils.SeekEnd(file)
	return core_utils.LogFuncBind(buffer, err, func(b io.ReadSeeker) (io.ReadSeeker, error) {
		return core.Hello(b, numLines, 64000)
	})
}

func ReadReverseNLines(filename string, numLines uint64) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := core_utils.SeekEnd(file)
	return core_utils.LogFuncBind(buffer, err, func(b io.ReadSeeker) (io.ReadSeeker, error) {
		return core.ReadReverseNLines(b, numLines)
	})
}

func ReadReversePassesFilter(filename string, expr string) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := core_utils.SeekEnd(file)
	return core_utils.LogFuncBind(buffer, err, func(b io.ReadSeeker) (io.ReadSeeker, error) {
		return core.ReadReversePassesFilter(b, expr)
	})
}
