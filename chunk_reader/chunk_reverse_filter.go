package chunk_reader

import (
	"bytes"
	"errors"
	"io"
	"log_monitor/monitor/core"
)

func ReadReversePassesFilter(reader io.ReadSeeker, expr string, chunk int64) (io.ReadSeeker, error) {
	validBlockCount := uint64(0)

	results := make(chan parseResult)
	expected := make(chan uint64)
	accumulated := make(chan io.ReadSeeker)
	errChannel := make(chan error)
	go AccumulateResults(results, expected, accumulated, errChannel)

	defer close(results)
	defer close(expected)

	var lastBlock parseBlock
	filter := GetReadReverseAsyncFuncFilter(results, expr)
	processBlock := GetProcessBlockReverseFunc(&lastBlock, func(index uint64, block parseBlock) {
		if(block.main != nil) {
			filter(validBlockCount, block.main, block.mainCount)
			validBlockCount++
		}
	})
	keepReading := func () bool {
		// early kill here?
		return true
	}

	i, err := ChunkRead(reader, chunk, ReadBackward, processBlock, keepReading)
	if err != nil {
		return nil, err
	}

	{
		dummy := parseBlock{ prefix: []byte("dummy\n")}
		dummy = stitchOtherBlockPrefix(dummy, lastBlock)
		if dummy.main != nil {
			processBlock(dummy.main, len(dummy.main), i + 1)
			i++
		} else {
			return nil, errors.New("parse error")
		}
	}

	expected <- validBlockCount

	err = <- errChannel
	if err != nil {
		return nil, err
	}
	return <- accumulated, nil
}

func GetReadReverseAsyncFuncFilter(parseResultChan chan<- parseResult, expr string) func(uint64, []byte, uint64) {
	return func(index uint64, buffer []byte, nLines uint64) {
		go func() {
			reader := bytes.NewReader(buffer)
			reader.Seek(0, io.SeekEnd)
			res, err := core.ReadReversePassesFilterFast(reader, expr)
		parseResultChan <- parseResult{
				index: index,
				result: res,
				err: err,
			}
		}()
	}
}

