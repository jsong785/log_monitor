package core_utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"log_monitor/monitor/test_utils"
	"strings"
	"testing"
)

func TestLogFuncMonad_ErrorAtStart(t *testing.T) {
	reader := strings.NewReader("init")
	res, err := LogFuncBind(reader, errors.New("init"), chainBuffer, chainAbc)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "init", err.Error())
}

func TestLogFuncMonad_ErrorAtFirstchain(t *testing.T) {
	reader := strings.NewReader("init")
	res, err := LogFuncBind(reader, nil, chainBufferError, chainAbc)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "buffer_error", err.Error())
}

func TestLogFuncMonad_ErrorAtSecondchain(t *testing.T) {
	reader := strings.NewReader("init")
	res, err := LogFuncBind(reader, nil, chainBuffer, chainAbcError)
	assert.NotNil(t, res) // last may get to return something
	assert.Equal(t, "abc", test_utils.GetString(res))
	assert.NotNil(t, err)
	assert.Equal(t, "abc_error", err.Error())
}

func TestLogFuncMonad_NoError(t *testing.T) {
	reader := strings.NewReader("init")
	res, err := LogFuncBind(reader, nil, chainBuffer, chainAbc)

	assert.NotNil(t, res)
	assert.Equal(t, "abc", test_utils.GetString(res))
	assert.Nil(t, err)
}

type mockSeek struct{}

func (*mockSeek) Read([]byte) (int, error) {
	return 0, nil
}
func (*mockSeek) Seek(int64, int) (int64, error) {
	return -1, errors.New("mock seek error")
}

func TestLogFuncMonad_SeekErrors(t *testing.T) {
	var mock mockSeek
	res, err := LogFuncBind(&mock, nil, chainBuffer, chainAbc)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "mock seek error", err.Error())
}

func chainBuffer(io.ReadSeeker) (io.ReadSeeker, error) {
	return strings.NewReader("buffer"), nil
}

func chainBufferError(io.ReadSeeker) (io.ReadSeeker, error) {
	return strings.NewReader("buffer"), errors.New("buffer_error")
}

func chainAbc(io.ReadSeeker) (io.ReadSeeker, error) {
	return strings.NewReader("abc"), nil
}

func chainAbcError(io.ReadSeeker) (io.ReadSeeker, error) {
	return strings.NewReader("abc"), errors.New("abc_error")
}
