package chunk_reader

import (
	"bytes"
	"errors"
	"io"
	"log_monitor/monitor/core"
)

func ReadReverseNLines(reader io.ReadSeeker, nLines uint64, chunk int64) (io.ReadSeeker, error) {
	count := uint64(0)
	validBlockCount := uint64(0)

	results := make(chan parseResult)
	expected := make(chan uint64)
	accumulated := make(chan io.ReadSeeker)
	errChannel := make(chan error)
	go AccumulateResults(results, expected, accumulated, errChannel)

	defer close(results)
	defer close(expected)

	var lastBlock parseBlock
	processBlock := GetProcessBlockReverseFunc(&lastBlock, GetProcessBlockReverseNLinesLimitFunc(&validBlockCount , &count, nLines, GetReadReverseNLinesAsyncFunc(results)))
	keepReading := func () bool {
		// early kill here?
		return count < nLines
	}

	i, err := ChunkRead(reader, chunk, ReadBackward, processBlock, keepReading)
	if err != nil {
		return nil, err
	}

	if count < nLines {
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

func GetProcessBlockReverseNLinesLimitFunc(index *uint64, currentCount *uint64, lineLimit uint64, processFunc func(uint64, []byte, uint64)) func(uint64, parseBlock) {
	return func(ba uint64, block parseBlock) {
		if(block.main != nil) {
			linesToProcess := block.mainCount
			if linesToProcess + *currentCount > lineLimit {
				linesToProcess = lineLimit - *currentCount
			}
			*currentCount += linesToProcess
			processFunc(*index, block.main, linesToProcess)
			*index++
		}
	}
}

func GetReadReverseNLinesAsyncFunc(parseResultChan chan<- parseResult) func(uint64, []byte, uint64) {
	return func(index uint64, buffer []byte, nLines uint64) {
		go func() {
			reader := bytes.NewReader(buffer)
			reader.Seek(0, io.SeekEnd)
			res, err := core.ReadReverseNLinesFast(reader, nLines)
		parseResultChan <- parseResult{
				index: index,
				result: res,
				err: err,
			}
		}()
	}
}
