package core_utils

import (
	"github.com/stretchr/testify/assert"
        "io"
        "strings"
	"testing"
)

func TestSeekEnd(t *testing.T) {
    reader := strings.NewReader("abc")

    pos, err := reader.Seek(0, io.SeekCurrent)
    assert.Nil(t, err)
    assert.Equal(t, int64(0), pos)

    ret, err := SeekEnd(reader)
    assert.Nil(t, err)
    assert.Equal(t, reader, ret)

    pos, err = ret.Seek(0, io.SeekCurrent)
    assert.Nil(t, err)
    assert.Equal(t, int64(3), pos)
}

