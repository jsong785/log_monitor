package file_reader

import (
	"io"
	"log_monitor/monitor/chunk_reader"
	"log_monitor/monitor/core"
	"log_monitor/monitor/core_utils"
	"os"
)

const chunkSize = int64(64000)

func ReadReverseNLinesChunk(filename string, numLines uint64) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := core_utils.SeekEnd(file)
	return core_utils.LogFuncBind(buffer, err, func(b io.ReadSeeker) (io.ReadSeeker, error) {
		return chunk_reader.ReadReverseNLines(b, numLines, chunkSize)
	})
}

func ReadReversePassesFilterChunk(filename string, expr string) (io.ReadSeeker, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer, err := core_utils.SeekEnd(file)
	return core_utils.LogFuncBind(buffer, err, func(b io.ReadSeeker) (io.ReadSeeker, error) {
		return chunk_reader.ReadReversePassesFilter(b, expr, chunkSize)
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
