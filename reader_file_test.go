package log_monitor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestReadLastNLines_File_small(t *testing.T) {
        line, err := ReadLastNLinesFromFile("syslog_ex", 2)
	assert.Nil(t, err)
	assert.Equal(t, []string{"jkl\n", "ghi\n"}, line)
}

func TestReadLastLinesContainsString_File_small(t *testing.T) {
        line, err := ReadLastLinesContainsStringFromFile("syslog_ex", "_")
	assert.Nil(t, err)
	assert.Equal(t, []string{ "_world\n", "_hello\n"}, line)
}

