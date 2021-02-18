package log_monitor

import (
	"github.com/stretchr/testify/assert"
        "sync"
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

func BenchmarkLargeFile(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ReadLastNLinesFromFile("syslog_large", 1000)
    }
}

func BenchmarkLoadTest(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        for j := 0; j < 10000; j++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                _, err := ReadLastNLinesFromFile("syslog_large", 20)
                assert.Nil(b, err)
            }()
        }
        wg.Wait()
    }
}

