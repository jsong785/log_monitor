package chunk_reader

import (
	"bytes"
	"errors"
	"io"
	"log_monitor/monitor/core"
	"sort"
)

func ReadReversePassesFilter(buffer io.ReadSeeker, expr string, chunk int64) (io.ReadSeeker, error) {
	parseBlockFunc := func(block parseBlock, index uint64, resultsChan chan<- parseResult) {
		reader := bytes.NewReader(block.main)
		reader.Seek(0, io.SeekEnd)
		res, err := core.ReadReversePassesFilterFast(reader, expr)
		if err != nil {
			//fmt.Println(err)
		}
		resultsChan <- parseResult{
			index:  index,
			result: res,
			err:    err,
		}

	}
	processedLinesFunc := func(uint64) {
	}
	keepGoingFunc := func() bool {
		return true
	}
	return chunkReader(buffer, parseBlockFunc, processedLinesFunc, keepGoingFunc, chunk)
}

func ReadReverseNLines(buffer io.ReadSeeker, nLines uint64, chunk int64) (io.ReadSeeker, error) {
	lines := uint64(0)
	parseBlockFunc := func(block parseBlock, index uint64, resultsChan chan<- parseResult) {
		reader := bytes.NewReader(block.main)
		reader.Seek(0, io.SeekEnd)

		linesToProcess := block.mainCount
		if lines >= nLines {
			linesToProcess -= (lines - nLines)
		}

		res, err := core.ReadReverseNLinesFast(reader, linesToProcess)
		if err != nil {
			//fmt.Println(err)
		}
		resultsChan <- parseResult{
			index:  index,
			result: res,
			err:    err,
		}

	}
	processedLinesFunc := func(count uint64) {
		lines = count
	}
	keepGoingFunc := func() bool {
		return lines < nLines
	}
	return chunkReader(buffer, parseBlockFunc, processedLinesFunc, keepGoingFunc, chunk)
}

type parseResult struct {
	index  uint64
	result io.ReadSeeker
	err    error
}
type byIndex []parseResult

func (a byIndex) Len() int { return len(a) }
func (a byIndex) Swap(i, j int) {
	cache := a[i]
	cache2 := a[j]
	a[j] = cache
	a[i] = cache2
}
func (a byIndex) Less(i, j int) bool { return a[i].index < a[j].index }

// assumes chunk is above 0
func readChunk(buffer io.ReadSeeker, chunk int64) (parseBlock, uint64, error) {
	currentPos, err := buffer.Seek(0, io.SeekCurrent)
	if err != nil {
		return parseBlock{}, 0, err
	}

	seekAmt := -chunk
	if currentPos < chunk {
		seekAmt = -currentPos
	}
	if seekAmt == 0 {
		return parseBlock{}, 0, nil
	}

	_, err = buffer.Seek(seekAmt, io.SeekCurrent)
	if err != nil {
		return parseBlock{}, 0, err
	}

	read := make([]byte, -seekAmt)
	amt, err := buffer.Read(read)
	if err != nil {
		return parseBlock{}, 0, err
	} else if amt != len(read) {
		return parseBlock{}, 0, errors.New("truncation detected")
	}
	_, err = buffer.Seek(seekAmt, io.SeekCurrent)
	if err != nil {
		return parseBlock{}, 0, err
	}
	return getParseBlock(read), uint64(amt), nil
}

func chunkReader(buffer io.ReadSeeker, parseBlockFunc func(parseBlock, uint64, chan<- parseResult), processed func(uint64), keepGoing func() bool, chunk int64) (io.ReadSeeker, error) {
	if chunk <= 0 {
		return nil, errors.New("cache size must be above zero")
	}

	resultsChannel := make(chan parseResult)
	blockAndUpdateSlice := func(parseResults []parseResult) []parseResult {
		select {
		case res := <-resultsChannel:
			parseResults = append(parseResults, res)
		default:
			break
		}
		return parseResults
	}

	resultsSlice := make([]parseResult, 0)

	var previousBlock parseBlock
	linesProcessedAsChunks := uint64(0)
	forceBreak := false

	parseBlockIndex := uint64(0)
	for keepGoing() && !forceBreak {
		block, read, err := readChunk(buffer, chunk)
		if err != nil {
			return nil, err
			// if this dies, the go threads might still be running, figure out later
		}
		if read == 0 { // managed to read everything to 0 perfectly aligned
			break
		} else if read < uint64(chunk) { // managed to read to start of file (not aligned with chunk)
			forceBreak = true // last
		}

		block = stitchOtherBlockPrefix(block, previousBlock)
		previousBlock = block
		if forceBreak && block.prefix != nil {
			block.main = append(block.prefix, block.main...)
			block.mainCount++
		}
		if block.main == nil {
			continue
		}
		linesProcessedAsChunks += block.mainCount
		processed(linesProcessedAsChunks)
		go func() {
			parseBlockFunc(block, parseBlockIndex, resultsChannel)
		}()
		parseBlockIndex++
		resultsSlice = blockAndUpdateSlice(resultsSlice)
	}
	for uint64(len(resultsSlice)) < parseBlockIndex {
		select {
		case res := <-resultsChannel:
			resultsSlice = append(resultsSlice, res)
		default:
			break
		}
	}
	sort.Sort(byIndex(resultsSlice))

	var results bytes.Buffer
	for _, res := range resultsSlice {
		if res.err != nil {
			return nil, res.err
		}
		_, err := results.ReadFrom(res.result) // may overfllow
		if err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(results.Bytes()), nil
}

type parseBlock struct {
	prefix    []byte
	main      []byte
	suffix    []byte
	mainCount uint64
}

func getParseBlock(buffer []byte) parseBlock {
	var block parseBlock
	if len(buffer) == 0 {
		return block
	}

	mainCount := uint64(0)
	lastNewLineIndex := -1
	for index, c := range buffer {
		if c == '\n' {
			if len(block.prefix) == 0 {
				block.prefix = buffer[:index+1]
			} else {
				mainCount++
			}
			lastNewLineIndex = index
		}
	}
	if lastNewLineIndex == -1 {
		block.suffix = buffer
		return block
	} else if len(block.prefix)-1 == lastNewLineIndex && len(buffer)-1 == lastNewLineIndex {
		return block
	}

	if len(block.prefix)-1 != lastNewLineIndex {
		block.main = buffer[len(block.prefix) : lastNewLineIndex+1]
		block.mainCount = mainCount
	}

	if len(block.prefix)+len(block.main) == len(buffer) {
		return block
	}

	block.suffix = buffer[lastNewLineIndex+1:]
	return block
}

func stitchOtherBlockPrefix(one parseBlock, two parseBlock) parseBlock {
	var ret parseBlock

	ret.prefix = one.prefix
	ret.main = one.main
	ret.mainCount = one.mainCount
	if len(two.prefix) > 0 {
		var other []byte
		other = append(other, one.suffix...)
		other = append(other, two.prefix...)
		if one.prefix == nil {
			ret.prefix = append(ret.prefix, other...)
		} else {
			ret.main = append(ret.main, other...)
			ret.mainCount = one.mainCount + 1
		}
	}
	return ret
}
