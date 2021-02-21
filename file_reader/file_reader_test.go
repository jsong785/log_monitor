package file_reader

import (
    "os"
    "io"
	"github.com/stretchr/testify/assert"
	"log_monitor/monitor/test_utils"
	"sync"
	"testing"
)

func TestReadLastNLines_File_small(t *testing.T) {
	line, err := ReadReverseNLines("../files/syslog_ex", 2)
	assert.Nil(t, err)
	assert.Equal(t, []string{"jkl\n", "ghi\n"}, test_utils.GetLines(line))
}

func TestReadLastLinesContainsString_File_small(t *testing.T) {
	line, err := ReadReversePassesFilter("../files/syslog_ex", "_")
	assert.Nil(t, err)
	assert.Equal(t, []string{"_world\n", "_hello\n"}, test_utils.GetLines(line))
}

func BenchmarkLargeFile_SingleRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ReadReverseNLines("../files/syslog_large", 1000)
		assert.Nil(b, err)
	}
}

func TestEqualBleh(t *testing.T) {
    a, _ := ReadReverseNLines("../files/syslog_large", 1000)
    b, _ := PocReadReverseNLines("../files/syslog_large", 1000)
    af, _ := os.Create("AAA")
    defer af.Close()
    io.Copy(af, a)

    bf, _ := os.Create("BBB")
    defer bf.Close()
    io.Copy(bf, b)
}

func BenchmarkLargeFile_SingleRequestPoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := PocReadReverseNLines("../files/syslog_large", 100000)
		assert.Nil(b, err)
	}
}

func BenchmarkLargeFile_ManyRequests(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 1000; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := PocReadReverseNLines("../files/syslog_large", 1000)
				assert.Nil(b, err)
			}()
		}
		wg.Wait()
	}
}
