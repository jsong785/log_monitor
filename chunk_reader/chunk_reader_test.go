package chunk_reader

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log_monitor/monitor/test_utils"
	"io"
	"strings"
	"testing"
)

func TestReadReverseN(t *testing.T) {
	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("abc\ndef\nghi\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReverseNLines(reader, 3, 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ghi\n", "def\n", "abc\n"}, test_utils.GetLines(res))
	});

	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("abc\ndef\nghi\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReverseNLines(reader, 3, 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ghi\n", "def\n", "abc\n"}, test_utils.GetLines(res))
	});

	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("abc\ndef\nghi\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReverseNLines(reader, 3, 3)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ghi\n", "def\n", "abc\n"}, test_utils.GetLines(res))
	});

	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("abc\ndef\nghi\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReverseNLines(reader, 3, 10000)
		assert.Nil(t, err)
		assert.Equal(t, []string{"ghi\n", "def\n", "abc\n"}, test_utils.GetLines(res))
	});
}

func TestReadReverseNFilter(t *testing.T) {
	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("abc\ndef\nghi\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReversePassesFilter(reader, "e", 1)
		assert.Nil(t, err)
		assert.Equal(t, []string{"def\n"}, test_utils.GetLines(res))
	});

	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("aob\ncde\nfog\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReversePassesFilter(reader, "o", 2)
		assert.Nil(t, err)
		assert.Equal(t, []string{"fog\n", "aob\n"}, test_utils.GetLines(res))
	});

	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("aob\ncde\nfog\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReversePassesFilter(reader, "o", 3)
		assert.Nil(t, err)
		assert.Equal(t, []string{"fog\n", "aob\n"}, test_utils.GetLines(res))
	});

	t.Run("ran", func(t *testing.T) {
		reader := strings.NewReader("aob\ncde\nfog\n")
		reader.Seek(0, io.SeekEnd)
		res, err := ReadReversePassesFilter(reader, "o", 10000)
		assert.Nil(t, err)
		assert.Equal(t, []string{"fog\n", "aob\n"}, test_utils.GetLines(res))
	});
}

func TestAccumulatedResults(t *testing.T) {
	t.Run("", func(t *testing.T){
		results := make(chan parseResult)
		expected := make(chan uint64)
		accumulated := make(chan io.ReadSeeker)
		err := make(chan error)

		go AccumulateResults(results, expected, accumulated, err)
	})
}

func TestReadReverseAsync(t *testing.T) {
	c := make(chan parseResult)
	f := GetReadReverseAsyncFunc(c)

	f(0, []byte("abc\ndef\ngef\n"), 2)
	res := <- c
	assert.Nil(t, res.err)
	assert.Equal(t, uint64(0), res.index)
	assert.Equal(t, []string{"gef\n", "def\n"}, test_utils.GetLines(res.result))

	f(1, []byte("abc\ndef\ngef\n"), 5)
	res = <- c
	assert.NotNil(t, res.err)
	assert.Equal(t, uint64(1), res.index)
	assert.Nil(t, res.result)
}

func TestGetProcessBlockReverseNLinesLimitFunc(t *testing.T) {
	var buffer []byte
	index := uint64(0)
	lines := uint64(0)
	process := func(i uint64, b []byte, n uint64) {
		index = i
		buffer = b
		lines = n
	}

	validBlockCount := uint64(0)
	t.Run("limit is 2, parse 2 lines, then try to parse another 2", func(t *testing.T) {
		current := uint64(0)

		f := GetProcessBlockReverseNLinesLimitFunc(&validBlockCount, &current, 2, process)
		f(0, CreateBlockWithCount("123\n", "456\n789\n", "", 2))
		assert.Equal(t, uint64(0), index)
		assert.Equal(t, uint64(2), lines)
		assert.Equal(t,"456\n789\n", string(buffer))
		assert.Equal(t, uint64(1), validBlockCount)

		f(1, CreateBlockWithCount("123\n", "456\n789\n", "", 2))
		assert.Equal(t, uint64(1), index)
		assert.Equal(t, uint64(0), lines)
		assert.Equal(t, uint64(2), validBlockCount)
	})

	validBlockCount = 0
	t.Run("limit is 2, parse block has 3 valid lines", func(t *testing.T) {
		current := uint64(0)

		f := GetProcessBlockReverseNLinesLimitFunc(&validBlockCount, &current, 2, process)
		f(0, CreateBlockWithCount("123\n", "456\n789\nabc\n", "", 3))
		assert.Equal(t, uint64(0), index)
		assert.Equal(t, uint64(2), lines)
		assert.Equal(t,"456\n789\nabc\n", string(buffer))
		assert.Equal(t, uint64(1), validBlockCount)

		f(1, CreateBlockWithCount("123\n", "456\n789\nabc\n", "", 3))
		assert.Equal(t, uint64(1), index)
		assert.Equal(t, uint64(0), lines)
		assert.Equal(t,"456\n789\nabc\n", string(buffer))
		assert.Equal(t, uint64(2), validBlockCount)
	})
}

func TestGetProcessReverseBlockFunc(t *testing.T) {
	var index uint64
	var last parseBlock
	blockFunc := func(i uint64, b parseBlock) {
		index = i
		last = b
	}

	f := GetProcessBlockReverseFunc(blockFunc)

	f([]byte("123\n"), 4, 0)
	assert.Equal(t, uint64(0), index)
	assert.Equal(t, CreateBlock("123\n", "", ""), last)

	f([]byte("456\n789\n"), 4, 1)
	assert.Equal(t, uint64(1), index)
	assert.Equal(t, CreateBlock("456\n", "123\n", ""), last)

	f([]byte("abc\ndef\n"), 8, 2)
	assert.Equal(t, uint64(2), index)
	assert.Equal(t, CreateBlockWithCount("abc\n", "def\n456\n", "", 2), last)
}

func TestgetReadOfset(t *testing.T) {
	t.Run("forwards", func(t *testing.T) {
		assert.Equal(t, int64(0), getReadOffset(ReadForward, 0, 0))
		assert.Equal(t, int64(1000), getReadOffset(ReadForward, 1000, 0))
	})
	t.Run("backwards", func(t *testing.T) {
		assert.Equal(t, int64(0), getReadOffset(ReadForward, 0, 0))
		assert.Equal(t, int64(-1000), getReadOffset(ReadForward, 1000, 0))
		assert.Equal(t, int64(-500), getReadOffset(ReadForward, 500, 0)) // don't seek back past current
	})
}

func TestChunkReadForwards(t *testing.T) {
	t.Run("forwards - invalid buffer size", func(t *testing.T) {
		reader := strings.NewReader("")
		i, err := ChunkRead(reader, -1, ReadForward, func([]byte, int, uint64){}, func() bool{ return true })
		assert.Equal(t, uint64(0), i)
		assert.Equal(t, "cache size must be above zero", err.Error())

		i, err =  ChunkRead(reader, 0, ReadForward, func([]byte, int, uint64){}, func() bool{ return true })
		assert.Equal(t, uint64(0), i)
		assert.Equal(t, "cache size must be above zero", err.Error())
	})
	t.Run("forwards - empty", func(t *testing.T) {
		reader := strings.NewReader("")
		i, err := ChunkRead(reader, 1, ReadForward, func([]byte, int, uint64){}, func() bool{ return true })
		assert.Equal(t, uint64(0), i)
		assert.Nil(t, err)

		i, err = ChunkRead(reader, 2, ReadForward, func([]byte, int, uint64){}, func() bool{ return true })
		assert.Equal(t, uint64(0), i)
		assert.Nil(t, err)
	})
	t.Run("forwards - filled buffer - by 1", func(t *testing.T) {
		reader := strings.NewReader("123456")

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err := ChunkRead(reader, 1, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				index = append(index, idx)
				buffer = append(buffer, b[:amt]...)
			}, func() bool{ return true })

		assert.Equal(t, uint64(5), i)
		assert.Nil(t, err)
		assert.Equal(t, []uint64{ 0, 1, 2, 3, 4, 5 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(6), pos)
	})
	t.Run("forwards - filled buffer - by 3", func(t *testing.T) {
		reader := strings.NewReader("123456")

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err := ChunkRead(reader, 3, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				index = append(index, idx)
				buffer = append(buffer, b[:amt]...)
			}, func() bool{ return true })

		assert.Equal(t, uint64(1), i)
		assert.Nil(t, err)
		assert.Equal(t, []uint64{ 0, 1 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(6), pos)
	})
	t.Run("forwards - filled buffer - by 5", func(t *testing.T) {
		reader := strings.NewReader("123456")

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err :=  ChunkRead(reader, 5, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return true })

		assert.Equal(t, uint64(1), i)
		assert.Nil(t, err)
		assert.Equal(t, []uint64{ 0, 1 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(6), pos)
	})
	t.Run("forwards - filled buffer - by 6", func(t *testing.T) {
		reader := strings.NewReader("123456")

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err :=  ChunkRead(reader, 6, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return true })

		assert.Equal(t, uint64(0), i)
		assert.Nil(t, err)
		assert.Equal(t, []uint64{ 0 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(6), pos)
	})
	t.Run("forwards - filled buffer - by 8", func(t *testing.T) {
		reader := strings.NewReader("123456")

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err :=  ChunkRead(reader, 8, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return true })

		assert.Equal(t, uint64(0), i)
		assert.Nil(t, err)
		assert.Equal(t, []uint64{ 0 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(6), pos)
	})
	t.Run("forwards - filled buffer - by 3 - terminate before first loop", func(t *testing.T) {
		reader := strings.NewReader("123456")

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err :=  ChunkRead(reader, 3, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return false })

		assert.Equal(t, uint64(0), i)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(index))
		assert.Equal(t, 0, len(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("forwards - filled buffer - by 3, terminate early", func(t *testing.T) {
		reader := strings.NewReader("123456")

		keepReading := true
		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		i, err :=  ChunkRead(reader, 3, ReadForward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ 
				cache := keepReading
				keepReading = !keepReading
				return cache
			})

		assert.Equal(t, uint64(0), i)
		assert.Nil(t, err)
		assert.Equal(t, []uint64{ 0 }, index)
		assert.Equal(t, "123", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(3), pos)
	})
}

/*
func TestChunkReadBackwards(t *testing.T) {
	t.Run("backwards - invalid buffer size", func(t *testing.T) {
		reader := strings.NewReader("")
		reader.Seek(0, io.SeekEnd)
		assert.Equal(t, "cache size must be above zero", ChunkRead(reader, -1, ReadBackward, func([]byte, int, uint64){}, func() bool{ return true }).Error())
		assert.Equal(t, "cache size must be above zero", ChunkRead(reader, 0, ReadBackward, func([]byte, int, uint64){}, func() bool{ return true }).Error())
	})
	t.Run("backwards - empty", func(t *testing.T) {
		reader := strings.NewReader("")
		reader.Seek(0, io.SeekEnd)
		assert.Nil(t, ChunkRead(reader, 1, ReadBackward, func([]byte, int, uint64){}, func() bool{ return true }))
		assert.Nil(t, ChunkRead(reader, 2, ReadBackward, func([]byte, int, uint64){}, func() bool{ return true }))
	})
	t.Run("backwards - filled buffer - by 1", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 1, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				index = append(index, idx)
				buffer = append(buffer, b[:amt]...)
			}, func() bool{ return true }))

		assert.Equal(t, []uint64{ 0, 1, 2, 3, 4, 5}, index)
		assert.Equal(t, "654321", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("backwards - filled buffer - by 3", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 3, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return true }))

		assert.Equal(t, []uint64{ 0, 1 }, index)
		assert.Equal(t, "456123", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("backwards - filled buffer - by 5", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 5, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				index = append(index, idx)
				buffer = append(buffer, b[:amt]...)
			}, func() bool{ return true }))

		assert.Equal(t, []uint64{ 0, 1 }, index)
		assert.Equal(t, "234561", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("backwards - filled buffer - by 6", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 6, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return true }))

		assert.Equal(t, []uint64{ 0 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("backwards - filled buffer - by 8", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 8, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return true }))

		assert.Equal(t, []uint64{ 0 }, index)
		assert.Equal(t, "123456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("backwards - filled buffer - by 3 - terminate before first loop", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 3, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{ return false }))

		assert.Equal(t, 0, len(index))
		assert.Equal(t, 0, len(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(6), pos)
	})
	t.Run("backwards - filled buffer - by 3  - terminate early", func(t *testing.T) {
		reader := strings.NewReader("123456")
		reader.Seek(0, io.SeekEnd)

		keepReading := true
		index := make([]uint64, 0)
		buffer := make([]byte, 0)
		assert.Nil(t, ChunkRead(reader, 3, ReadBackward, 
			func(b []byte, amt int, idx uint64) {
				buffer = append(buffer, b[:amt]...)
				index = append(index, idx)
			}, func() bool{
				c := keepReading
				keepReading = !keepReading
				return c
			}))

		assert.Equal(t, []uint64{ 0 }, index)
		assert.Equal(t, "456", string(buffer))
		
		pos, _ := reader.Seek(0, io.SeekCurrent)
		assert.Equal(t, int64(3), pos)
	})
}
*/

func TestReadNLines(t *testing.T) {
	reader := strings.NewReader("123\n456\n789\n")
	reader.Seek(0, io.SeekEnd)
	res, err := ReadReverseNLines(reader, 3, 100)
	//_ = res
	//assert.Nil(t, err)
	assert.Nil(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(res)
	assert.Equal(t, "789\n456\n123\n", buf.String())
}

func TeststitchOtherBlockPrefix(t *testing.T) {
	// stitch two full blocks
	func() {
		one := CreateBlock("a", "b", "c")
		two := CreateBlock("d", "e", "f")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("bcd"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(2), stitched.mainCount)
	}()

	// stitch two empty
	func() {
		one := CreateBlock("", "", "")
		two := CreateBlock("", "", "")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Nil(t, stitched.prefix)
		assert.Nil(t, stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(0), stitched.mainCount)
	}()

	// stitch two empty blocks with prefixes
	func() {
		one := CreateBlock("a", "", "")
		two := CreateBlock("b", "", "")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("b"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(1), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "", "")
		two := CreateBlock("b", "c", "")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("b"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(1), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "", "")
		two := CreateBlock("b", "c", "d")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("b"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(1), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "x", "")
		two := CreateBlock("b", "c", "d")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("xb"), stitched.main)
		assert.Equal(t, uint64(2), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "", "")
		two := CreateBlock("b", "c", "d")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("b"), stitched.main)
		assert.Equal(t, uint64(1), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "b", "c")
		two := CreateBlock("", "d", "e")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("b"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(1), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "", "")
		two := CreateBlock("", "", "b")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Nil(t, stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(0), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "b", "")
		two := CreateBlock("c", "d", "")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("bc"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(2), stitched.mainCount)
	}()

	func() {
		one := CreateBlock("a", "", "")
		two := CreateBlock("c", "d", "")

		stitched := stitchOtherBlockPrefix(one, two)
		assert.Equal(t, []byte("a"), stitched.prefix)
		assert.Equal(t, []byte("c"), stitched.main)
		assert.Nil(t, stitched.suffix)
		assert.Equal(t, uint64(1), stitched.mainCount)
	}()
}

func TestgetParseBlock(t *testing.T) {
	// empty string
	block := getParseBlock(bytes.NewBufferString("").Bytes())
	assert.Equal(t, block.mainCount, uint64(0))
	assert.Nil(t, block.prefix)
	assert.Nil(t, block.main)
	assert.Nil(t, block.suffix)

	// not a valid line
	block = getParseBlock(bytes.NewBufferString("123").Bytes())
	assert.Equal(t, block.mainCount, uint64(0))
	assert.Nil(t, block.prefix)
	assert.Nil(t, block.main)
	assert.Equal(t, []byte("123"), block.suffix)

	// blank line
	block = getParseBlock(bytes.NewBufferString("\n").Bytes())
	assert.Equal(t, block.mainCount, uint64(0))
	assert.Equal(t, []byte("\n"), block.prefix)
	assert.Nil(t, block.main)
	assert.Nil(t, block.suffix)

	// blank lines
	block = getParseBlock(bytes.NewBufferString("\n\n\n").Bytes())
	assert.Equal(t, block.mainCount, uint64(2))
	assert.Equal(t, []byte("\n"), block.prefix)
	assert.Equal(t, []byte("\n\n"), block.main)
	assert.Nil(t, block.suffix)

	// blank lines with remainder
	block = getParseBlock(bytes.NewBufferString("\n\n\n123").Bytes())
	assert.Equal(t, block.mainCount, uint64(2))
	assert.Equal(t, []byte("\n"), block.prefix)
	assert.Equal(t, []byte("\n\n"), block.main)
	assert.Equal(t, []byte("123"), block.suffix)

	// one line
	block = getParseBlock(bytes.NewBufferString("123\n").Bytes())
	assert.Equal(t, block.mainCount, uint64(0))
	assert.Equal(t, []byte("123\n"), block.prefix)
	assert.Nil(t, block.main)
	assert.Nil(t, block.suffix)

	// two lines
	block = getParseBlock(bytes.NewBufferString("123\n456\n").Bytes())
	assert.Equal(t, block.mainCount, uint64(1))
	assert.Equal(t, []byte("123\n"), block.prefix)
	assert.Equal(t, []byte("456\n"), block.main)
	assert.Nil(t, block.suffix)

	// three lines
	block = getParseBlock(bytes.NewBufferString("123\n456\n789\n").Bytes())
	assert.Equal(t, block.mainCount, uint64(2))
	assert.Equal(t, []byte("123\n"), block.prefix)
	assert.Equal(t, []byte("456\n789\n"), block.main)
	assert.Nil(t, block.suffix)

	// four lines
	block = getParseBlock(bytes.NewBufferString("123\n456\n789\n012\n").Bytes())
	assert.Equal(t, block.mainCount, uint64(3))
	assert.Equal(t, []byte("123\n"), block.prefix)
	assert.Equal(t, []byte("456\n789\n012\n"), block.main)
	assert.Nil(t, block.suffix)

	// four lines, partial 5th
	block = getParseBlock(bytes.NewBufferString("123\n456\n789\n012\nabc").Bytes())
	assert.Equal(t, block.mainCount, uint64(3))
	assert.Equal(t, []byte("123\n"), block.prefix)
	assert.Equal(t, []byte("456\n789\n012\n"), block.main)
	assert.Equal(t, []byte("abc"), block.suffix)

	// partial 2nd (one line)
	block = getParseBlock(bytes.NewBufferString("123\n4").Bytes())
	assert.Equal(t, block.mainCount, uint64(0))
	assert.Equal(t, []byte("123\n"), block.prefix)
	assert.Nil(t, block.main)
	assert.Equal(t, []byte("4"), block.suffix)
}

func CreateBlockWithCount(prefix string, main string, suffix string, mainCount uint64) parseBlock {
	var block parseBlock
	if len(prefix) > 0 {
		block.prefix = []byte(prefix)
	}
	if len(main) > 0 {
		block.main = []byte(main)
		block.mainCount = mainCount
	}
	if len(suffix) > 0 {
		block.suffix = []byte(suffix)
	}
	return block
}

func CreateBlock(prefix string, main string, suffix string) parseBlock {
	var block parseBlock
	if len(prefix) > 0 {
		block.prefix = []byte(prefix)
	}
	if len(main) > 0 {
		block.main = []byte(main)
		block.mainCount = 1
	}
	if len(suffix) > 0 {
		block.suffix = []byte(suffix)
	}
	return block
}
