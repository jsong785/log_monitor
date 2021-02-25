package chunk_reader

import (
	"bytes"
	"errors"
	"math"
	"io"
	"sort"
)

const ReadForward = 0
const ReadBackward = 1

func getReadOffset(direction int, chunk int64, position int64) int64 {
	if direction == ReadBackward {
		if chunk > position  {
			return -position
		}
		return -chunk
	} else {
		return chunk // even if past eof
	}
}

func GetProcessBlockReverseFunc(last *parseBlock, parseFunc func(uint64, parseBlock)) func ([]byte, int, uint64) {
	return func(buffer []byte, amt int, index uint64) {
		block := getParseBlock(buffer[:amt])
		block = stitchOtherBlockPrefix(block, *last)
		*last = block
		parseFunc(index, block)
	}
}

func AccumulateResults(results <-chan parseResult, expectedMaxIndex <-chan uint64, accumulated chan<- io.ReadSeeker, errorReport chan<- error) {
	bufferedResults := make([]parseResult, 0)
	checkMax := false
	max := uint64(0)
	for !checkMax || uint64(len(bufferedResults)) < max {
		select {
			case res := <-results:
			if res.err != nil { 
					close(accumulated)
					defer close(errorReport)
					errorReport <- res.err
					return
				}
				bufferedResults = append(bufferedResults, res)
			case max = <-expectedMaxIndex:
				checkMax = true
		}
	}

	sort.Sort(byIndex(bufferedResults))
	var buffer bytes.Buffer
	for _, r := range bufferedResults {
		_, err := buffer.ReadFrom(r.result)
		if err != nil {
			close(accumulated)
			defer close(errorReport)
			errorReport <- err
			return
		}
	}
	close(errorReport)
	defer close(accumulated)
	accumulated <- bytes.NewReader(buffer.Bytes())
}

func ChunkRead(reader io.ReadSeeker, chunk int64, direction int, processChunk func([]byte, int, uint64), keepReading func() bool) (uint64, error) {
	if chunk <= 0 {
		return 0, errors.New("cache size must be above zero")
	}

	somethingProcessed := false
	index := uint64(0)
	currentChunk := chunk
	for ; currentChunk == chunk && keepReading(); index++ {
		pos, err := reader.Seek(0, io.SeekCurrent)
		if err != nil {
			return index, err
		}

		offset := getReadOffset(direction, chunk, pos)
		if offset == 0 {
			break
		}
		currentChunk = int64(math.Abs(float64(offset)))

		if direction == ReadBackward {
			pos, err = reader.Seek(offset, io.SeekCurrent)
			if err != nil {
				return index, err
			}
		}

		buffer := make([]byte, int64(math.Abs(float64(offset))))
		amtRead, err := reader.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return index, err
			}
		} else if direction == ReadBackward && int64(amtRead) < -offset {
			return index, errors.New("truncation detected")
		}
		processChunk(buffer, amtRead, index)
		somethingProcessed = true

		if direction == ReadBackward {
			_, err = reader.Seek(offset, io.SeekCurrent)
			if err != nil {
				return index, err
			}
		}
	}
	if somethingProcessed {
		return index-1, nil
	}
	return index, nil
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
