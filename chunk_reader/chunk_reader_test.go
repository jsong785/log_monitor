package chunk_reader

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestreadChunk(t *testing.T) {
	// empty
	func() {
		reader := strings.NewReader("")
		reader.Seek(0, io.SeekEnd)
		block, amt, err := readChunk(reader, 1)
		assert.Equal(t, block, CreateBlock("", "", ""))
		assert.Equal(t, uint64(0), amt)
		assert.Nil(t, err)

		reader.Seek(0, io.SeekEnd)
		block, amt, err = readChunk(reader, 2)
		assert.Equal(t, block, CreateBlock("", "", ""))
		assert.Equal(t, uint64(0), amt)
		assert.Nil(t, err)
	}()

	// each chunk lands on perfect lines
	func() {
		reader := strings.NewReader("123\n456\n789\n")
		reader.Seek(0, io.SeekEnd)

		b1, amt, err := readChunk(reader, 4)
		assert.Equal(t, b1, CreateBlock("789\n", "", ""))
		assert.Equal(t, uint64(1), amt)
		assert.Nil(t, err)

		b2, amt, err := readChunk(reader, 4)
		assert.Equal(t, b2, CreateBlock("456\n", "", ""))
		assert.Equal(t, uint64(3), amt)
		assert.Nil(t, err)

		b3, amt, err := readChunk(reader, 4)
		assert.Equal(t, b3, CreateBlock("123\n", "", ""))
		assert.Equal(t, uint64(3), amt)
		assert.Nil(t, err)
	}()

	// each chunk does not land on perfect lines
	func() {
		reader := strings.NewReader("123\n4567\n7890123\n")
		reader.Seek(0, io.SeekEnd)

		b1, amt, err := readChunk(reader, 4)
		assert.Equal(t, b1, CreateBlock("123\n", "", ""))
		assert.Equal(t, uint64(4), amt)
		assert.Nil(t, err)

		b2, amt, err := readChunk(reader, 4)
		assert.Equal(t, b2, CreateBlock("7890", "", ""))
		assert.Equal(t, uint64(4), amt)
		assert.Nil(t, err)

		b3, amt, err := readChunk(reader, 4)
		assert.Equal(t, b3, CreateBlock("567\n", "", ""))
		assert.Equal(t, uint64(4), amt)
		assert.Nil(t, err)

		b4, amt, err := readChunk(reader, 4)
		assert.Equal(t, b4, CreateBlock("23\n", "", "4"))
		assert.Equal(t, uint64(4), amt)
		assert.Nil(t, err)

		b5, amt, err := readChunk(reader, 4)
		assert.Equal(t, b5, CreateBlock("1", "", ""))
		assert.Equal(t, uint64(1), amt)
		assert.Nil(t, err)
	}()
}

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
