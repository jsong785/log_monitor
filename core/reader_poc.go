package core

import (
    "bytes"
    "fmt"
    "io"
    "time"
)

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
        seekBack := int64(-2)
        if firstLoop {
            firstLoop = false
            seekBack = int64(-1)
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
        cache := make([]byte, 1)
        amt, err := buffer.Read(cache)
        if err != nil {
            return nil, err
        }

        if !firstChan {
            start := time.Now()
            res := <- resChan
            dur += time.Since(start)
            io.Copy(&totalResults, res) // err check!
            lastBlock = <- blockChan
            currentCount += <- countChan
        }
        firstChan = false

        go func(omg uint64, cc []byte, other ParseBlock, blockChan chan ParseBlock, resChan chan io.ReadSeeker, countChan chan uint64) {
            block := GetParseBlock(cc)
            block = StitchOtherBlockPrefix(block, other)
            nf := bytes.NewReader(block.main)
            nf.Seek(0, io.SeekEnd)
            if omg + block.mainCount > nLines {
                to := nLines - omg
                res, _ := ReadReverseNLines(nf, to) // err che
                resChan <- res
                blockChan <- block
                countChan <- block.mainCount
            } else {
                res, _ := ReadReverseNLines(nf, block.mainCount) // err che
                resChan <- res
                blockChan <- block
                countChan <- block.mainCount
            }
        }(currentCount, cache[:amt], lastBlock, blockChan, resChan, countChan)
    }
    fmt.Println("total waitime")
    fmt.Println(dur)
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
    count := one.mainCount

    var ret ParseBlock
    ret.prefix = one.prefix
    ret.main = one.main
    if len(two.prefix) > 0 {
        ret.main = append(ret.main, one.suffix...)
        ret.main = append(ret.main, two.prefix...)
        count++
    }
    ret.mainCount = count
    return ret
}

func StitchArgs(blocks ...ParseBlock) (bytes.Buffer, uint64) {
    return Stitch(blocks)
}

func Stitch(blocks []ParseBlock) (bytes.Buffer, uint64) {
    var buffer bytes.Buffer

    count := uint64(0)
    for i, b := range blocks {
        if i != 0 {
            buffer.Write(b.prefix)
        }
        buffer.Write(b.main)
        count += b.mainCount
        if i != len(blocks) -1 {
            buffer.Write(b.suffix)
            count++
        }
    }

    return buffer, count
}

