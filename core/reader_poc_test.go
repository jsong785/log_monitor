package core

import (
    "io"
	"github.com/stretchr/testify/assert"
        "bytes"
        "strings"
	"testing"
)

func CreateBlock(prefix string, main string, suffix string) ParseBlock {
        var block ParseBlock
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

func TestReadNLines(t *testing.T) {
    reader := strings.NewReader("123\n456\n789\n")
    reader.Seek(0, io.SeekEnd)
    res, err := Hello(reader, 1, 100)
    
    var buf bytes.Buffer
    buf.ReadFrom(res)
    assert.Equal(t, "789\n", buf.String())
    assert.Nil(t, err)
}

func TestReverse(t *testing.T) {
    assert.Nil(t, Reverse(nil))

    a := CreateBlock("a", "b", "c")
    b := CreateBlock("d", "e", "f")
    c := CreateBlock("g", "h", "i")
    assert.Equal(t, []ParseBlock{ c, b, a }, Reverse([]ParseBlock{a, b, c}))
    assert.NotEqual(t, []ParseBlock{ a, b, c }, Reverse([]ParseBlock{a, b, c}))

    d := CreateBlock("j", "k", "l")
    assert.Equal(t, []ParseBlock{ d, c, b, a }, Reverse([]ParseBlock{a, b, c,d }))
    assert.NotEqual(t, []ParseBlock{ a, b, c, d }, Reverse([]ParseBlock{a, b, c, d}))
}

func TestStitchOtherBlockPrefix(t *testing.T) {
    // stitch two full blocks
    func() {
        one := CreateBlock("a", "b", "c")
        two := CreateBlock("d", "e", "f")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("bcd"), stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(2), stitched.mainCount)
    }()

    // stitch two empty
    func() {
        one := CreateBlock("", "", "")
        two := CreateBlock("", "", "")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Nil(t, stitched.prefix)
        assert.Nil(t, stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(0), stitched.mainCount)
    }()

    // stitch two empty blocks with prefixes
    func() {
        one := CreateBlock("a", "", "")
        two := CreateBlock("b", "", "")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("b"), stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(1), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "", "")
        two := CreateBlock("b", "c", "")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("b"), stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(1), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "", "")
        two := CreateBlock("b", "c", "d")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("b"), stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(1), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "x", "")
        two := CreateBlock("b", "c", "d")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("xb"), stitched.main)
        assert.Equal(t, uint64(2), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "", "")
        two := CreateBlock("b", "c", "d")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("b"), stitched.main)
        assert.Equal(t, uint64(1), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "b", "c")
        two := CreateBlock("", "d", "e")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("b"), stitched.main)
        assert.Nil(t,  stitched.suffix)
        assert.Equal(t, uint64(1), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "", "")
        two := CreateBlock("", "", "b")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Nil(t, stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(0), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "b", "")
        two := CreateBlock("c", "d", "")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("bc"), stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(2), stitched.mainCount)
    }()

    func() {
        one := CreateBlock("a", "", "")
        two := CreateBlock("c", "d", "")

        stitched := StitchOtherBlockPrefix(one, two)
        assert.Equal(t, []byte("a"), stitched.prefix)
        assert.Equal(t, []byte("c"), stitched.main)
        assert.Nil(t, stitched.suffix)
        assert.Equal(t, uint64(1), stitched.mainCount)
    }()
}

func TestGetParseBlock(t *testing.T) {
    // empty string
    block := GetParseBlock(bytes.NewBufferString("").Bytes())
    assert.Equal(t, block.mainCount, uint64(0))
    assert.Nil(t, block.prefix)
    assert.Nil(t, block.main)
    assert.Nil(t, block.suffix)

    // not a valid line
    block = GetParseBlock(bytes.NewBufferString("123").Bytes())
    assert.Equal(t, block.mainCount, uint64(0))
    assert.Nil(t, block.prefix)
    assert.Nil(t, block.main)
    assert.Equal(t, []byte("123"), block.suffix)

    // blank line
    block = GetParseBlock(bytes.NewBufferString("\n").Bytes())
    assert.Equal(t, block.mainCount, uint64(0))
    assert.Equal(t, []byte("\n"), block.prefix)
    assert.Nil(t, block.main)
    assert.Nil(t, block.suffix)

    // blank lines
    block = GetParseBlock(bytes.NewBufferString("\n\n\n").Bytes())
    assert.Equal(t, block.mainCount, uint64(2))
    assert.Equal(t, []byte("\n"), block.prefix)
    assert.Equal(t, []byte("\n\n"), block.main)
    assert.Nil(t, block.suffix)

    // blank lines with remainder
    block = GetParseBlock(bytes.NewBufferString("\n\n\n123").Bytes())
    assert.Equal(t, block.mainCount, uint64(2))
    assert.Equal(t, []byte("\n"), block.prefix)
    assert.Equal(t, []byte("\n\n"), block.main)
    assert.Equal(t, []byte("123"), block.suffix)

    // one line
    block = GetParseBlock(bytes.NewBufferString("123\n").Bytes())
    assert.Equal(t, block.mainCount, uint64(0))
    assert.Equal(t, []byte("123\n"), block.prefix)
    assert.Nil(t, block.main)
    assert.Nil(t, block.suffix)

    // two lines
    block = GetParseBlock(bytes.NewBufferString("123\n456\n").Bytes())
    assert.Equal(t, block.mainCount, uint64(1))
    assert.Equal(t, []byte("123\n"), block.prefix)
    assert.Equal(t, []byte("456\n"), block.main)
    assert.Nil(t, block.suffix)

    // three lines
    block = GetParseBlock(bytes.NewBufferString("123\n456\n789\n").Bytes())
    assert.Equal(t, block.mainCount, uint64(2))
    assert.Equal(t, []byte("123\n"), block.prefix)
    assert.Equal(t, []byte("456\n789\n"), block.main)
    assert.Nil(t, block.suffix)

    // four lines
    block = GetParseBlock(bytes.NewBufferString("123\n456\n789\n012\n").Bytes())
    assert.Equal(t, block.mainCount, uint64(3))
    assert.Equal(t, []byte("123\n"), block.prefix)
    assert.Equal(t, []byte("456\n789\n012\n"), block.main)
    assert.Nil(t, block.suffix)

    // four lines, partial 5th
    block = GetParseBlock(bytes.NewBufferString("123\n456\n789\n012\nabc").Bytes())
    assert.Equal(t, block.mainCount, uint64(3))
    assert.Equal(t, []byte("123\n"), block.prefix)
    assert.Equal(t, []byte("456\n789\n012\n"), block.main)
    assert.Equal(t, []byte("abc"), block.suffix)

    // partial 2nd (one line)
    block = GetParseBlock(bytes.NewBufferString("123\n4").Bytes())
    assert.Equal(t, block.mainCount, uint64(0))
    assert.Equal(t, []byte("123\n"), block.prefix)
    assert.Nil(t, block.main)
    assert.Equal(t, []byte("4"), block.suffix)
}

