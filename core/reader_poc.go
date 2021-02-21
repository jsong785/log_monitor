package core

import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "sort"
    "time"
)

type ParseResult struct{
    index uint64
    result io.ReadSeeker
    err error
}
type ByIndex []ParseResult
func (a ByIndex) Len() int           { return len(a) }
func (a ByIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIndex) Less(i, j int) bool { return a[i].index < a[j].index }

func Hello(buffer io.ReadSeeker, nLines uint64, cacheSz int64) (io.ReadSeeker, error) {
    if(cacheSz <= 0) {
        return nil, errors.New("cache size must be above zero")
    }
    var totalResults bytes.Buffer

    resultsSlice := make([]ParseResult, 0)
    resultsChannel := make(chan ParseResult)

    firstLoop := true
    forceBreak := false
    lineCount := uint64(0)
    var lastBlock ParseBlock
    validBlockIndex := uint64(0)

    start := time.Now()
    for lineCount < nLines && !forceBreak {
        currentPos, err := buffer.Seek(0, io.SeekCurrent)
        if err != nil {
            return nil, err
        }

        seekBackAmt := 2*cacheSz
        readBufferSz := cacheSz
        if firstLoop {
            firstLoop = false
            seekBackAmt = cacheSz
        }
        if currentPos < readBufferSz {
            forceBreak = true
            readBufferSz = currentPos
            seekBackAmt = currentPos
        }

        currentPos, err = buffer.Seek(-seekBackAmt, io.SeekCurrent)
        if err != nil {
            return nil, errors.New("truncation detected")
        }

        readBuffer := make([]byte, readBufferSz)
        amt, err := buffer.Read(readBuffer)
        if err != nil {
            return nil, err
        } else if int64(amt) != readBufferSz {
            return nil, errors.New("truncation detected")
        }

        block := GetParseBlock(readBuffer)
        //fmt.Println("block diagnostics")
        //fmt.Println(block)
        //fmt.Printf("%s, %s, %s\n", string(block.prefix), string(block.main), string(block.suffix))
        //fmt.Println(lastBlock)
        //fmt.Printf("%s, %s, %s\n", string(lastBlock.prefix), string(lastBlock.main), string(lastBlock.suffix))
        block = StitchOtherBlockPrefix(block, lastBlock)
        lastBlock = block
        if block.main != nil {
            //fmt.Println("valid block diagnostics")
            //fmt.Println(block)
            //fmt.Printf("%s, %s, %s\n", string(block.prefix), string(block.main), string(block.suffix))
            //fmt.Println("main")
            //fmt.Println(block.mainCount)
            lineCount += block.mainCount
            linesToRead := block.mainCount
            if lineCount > nLines {
                linesToRead -= lineCount - nLines
            }
            //fmt.Println("wakasdf")
            //fmt.Println(linesToRead)
            go func(buf []byte, nLines uint64, index uint64) {
                reader := bytes.NewReader(buf)
                reader.Seek(0, io.SeekEnd)
                res, err := ReadReverseNLines(reader, nLines)
                //fmt.Println("results gotten")
                //fmt.Println(nLines)
                resultsChannel <- ParseResult{
                    index: index,
                    result: res,
                    err: err,
                }
            }(block.main, linesToRead, validBlockIndex)
            //fmt.Println("others")
            //fmt.Println(validBlockIndex)
            //fmt.Println(block.mainCount)
            validBlockIndex++
            //fmt.Println(string(block.main))
        }

        select {
        case res:= <- resultsChannel:
            resultsSlice = append(resultsSlice, res)
        default:
            break
        }
    }

    dur := time.Since(start)
    fmt.Println("slowness")
    fmt.Println(dur)

    //fmt.Println("slice len")
    //fmt.Println(len(resultsSlice))
    start = time.Now()
    for uint64(len(resultsSlice)) < validBlockIndex {
        res := <-resultsChannel
        resultsSlice = append(resultsSlice, res)
    }
    dur = time.Since(start)
    fmt.Println("slice time")
    fmt.Println(dur)

    start = time.Now()
    sort.Sort(ByIndex(resultsSlice))
    dur = time.Since(start)
    fmt.Println("sort time")
    fmt.Println(dur)

    start = time.Now()
    for _, res := range resultsSlice {
        if res.err != nil {
            return nil, res.err
        }
        _, err := totalResults.ReadFrom(res.result)
        if err != nil {
            return nil, err
        }
    }
    dur = time.Since(start)
    fmt.Println("read time sort asdf")
    fmt.Println(dur)

    start = time.Now()
    blahblah := bytes.NewReader(totalResults.Bytes())
    dur = time.Since(start)
    fmt.Println("reader build")
    fmt.Println(dur)
    return blahblah, nil
    //return bytes.NewReader(totalResults.Bytes()), nil
}

func PocReverseNLines(buffer io.ReadSeeker, nLines uint64) (io.ReadSeeker, error) {
    var totalResults bytes.Buffer

    var lastBlock ParseBlock
    startFound := false
    firstLoop := true
    firstChan := true

    var dur time.Duration

    resChan := make(chan io.ReadSeeker)
    blockChan := make(chan ParseBlock)
    countChan := make(chan uint64)

    currentCount := uint64(0)
    for currentCount <= nLines {
        seekBack := int64(-128000)
        if firstLoop {
            firstLoop = false
            seekBack = int64(-64000)
        }
        _, err := buffer.Seek(seekBack, io.SeekCurrent)
        if err != nil {
            if !startFound {
                startFound = true
                _, err = buffer.Seek(0, io.SeekStart)
                if err != nil {
                    return nil, err
                }
            } else {
                return nil, err
            }
        } 
        cache := make([]byte, 64000)
        amt, err := buffer.Read(cache)
        if err != nil {
            return nil, err
        }

        if !firstChan {
            start := time.Now()
            res := <- resChan
            dur += time.Since(start)
            if res != nil {
                _, err := io.Copy(&totalResults, res) // err check!
                if err != nil {
                    return nil, err
                }
            }
            lastBlock = <- blockChan
            currentCount += <- countChan
            if(currentCount >= nLines) {
                break;
            }
        }
        firstChan = false

        go func(omg uint64, cc []byte, other ParseBlock, blockChan chan ParseBlock, resChan chan io.ReadSeeker, countChan chan uint64) {
            block := GetParseBlock(cc)
            block = StitchOtherBlockPrefix(block, other)
            if block.main == nil {
                resChan <- nil
                blockChan <- block
                countChan <- block.mainCount
                return
            }
            nf := bytes.NewReader(block.main)
            nf.Seek(0, io.SeekEnd)
            if omg + block.mainCount > nLines {
                to := nLines - omg
                //fmt.Printf("this should be last")
                //fmt.Printf("what %d\n", to)
                res, _ := ReadReverseNLines(nf, to) // err che
                resChan <- res
                blockChan <- block
                countChan <- block.mainCount
            } else {
                //fmt.Printf("what %d\n", block.mainCount)
                res, _ := ReadReverseNLines(nf, block.mainCount) // err che
                resChan <- res
                blockChan <- block
                countChan <- block.mainCount
            }
        }(currentCount, cache[:amt], lastBlock, blockChan, resChan, countChan)
    }
    //fmt.Println("total waitime")
    //fmt.Println(dur)
    return bytes.NewReader(totalResults.Bytes()), nil
}

type ParseBlock struct {
    prefix []byte
    main []byte
    suffix []byte
    mainCount uint64
}

func GetParseBlock(buffer []byte) ParseBlock {
    var block ParseBlock
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

    if len(block.prefix) -1 != lastNewLineIndex {
        block.main = buffer[len(block.prefix):lastNewLineIndex+1]
        block.mainCount = mainCount
    }

    if len(block.prefix) + len(block.main) == len(buffer) {
        return block
    }

    block.suffix = buffer[lastNewLineIndex+1:]
    return block
}

func Reverse(blocks []ParseBlock) []ParseBlock {
    if len(blocks) == 0 {
        return blocks
    }
    start := 0
    end := len(blocks)-1
    for start < end {
        a := blocks[start]
        b := blocks[end]
        blocks[start] = b
        blocks[end] = a

        start++
        end--
    }
    return blocks
}

func StitchOtherBlockPrefix(one ParseBlock, two ParseBlock) ParseBlock {
    var ret ParseBlock

    ret.prefix = one.prefix
    ret.main = one.main
    ret.mainCount = one.mainCount
    if len(two.prefix) > 0 {
        var other []byte
        other = append(other, one.suffix...)
        other = append(other, two.prefix...)
        if(one.prefix == nil) {
            ret.prefix = append(ret.prefix, other...)
        } else {
            ret.main = append(ret.main, other...)
            ret.mainCount = one.mainCount + 1
        }
    }
    return ret
}
