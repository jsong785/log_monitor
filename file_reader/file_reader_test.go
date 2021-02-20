package file_reader

import (
	"github.com/stretchr/testify/assert"
	"log_monitor/monitor/test_utils"
	"sync"
	"testing"
)

func TestReadLastNLines_File_small(t *testing.T) {
	line, err := ReadReverseNLines("syslog_ex", 2)
	assert.Nil(t, err)
	assert.Equal(t, []string{"jkl\n", "ghi\n"}, test_utils.GetLines(line))
}

func TestReadLastLinesContainsString_File_small(t *testing.T) {
	line, err := ReadReversePassesFilter("syslog_ex", "_")
	assert.Nil(t, err)
	assert.Equal(t, []string{"_world\n", "_hello\n"}, test_utils.GetLines(line))
}

func BenchmarkLargeFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ReadReverseNLines("syslog_large", 1000)
		assert.Nil(b, err)
	}
}

func BenchmarkLargeFileLoadTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 1000; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := ReadReverseNLines("syslog_large", 1000)
				assert.Nil(b, err)
			}()
		}
		wg.Wait()
	}
}

func BenchmarkSmallFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ReadReverseNLines("syslog_ex", 3)
	}
}

func BenchmarkSmallFileLoadTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 1000; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := ReadReverseNLines("syslog_ex", 3)
				assert.Nil(b, err)
			}()
		}
		wg.Wait()
	}
}
