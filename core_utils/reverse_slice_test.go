package core_utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverseBytes(t *testing.T) {
	// empty
	func() {
		var b []byte
		assert.Nil(t, ReverseBytes(b))
		assert.Nil(t, b)
	}()

	// odd number of items
	func() {
		b := []byte{'a', 'b', 'c'}
		res := ReverseBytes(b[:])
		assert.Equal(t, b, res)
		assert.Equal(t, []byte{'c', 'b', 'a'}, b)
	}()

	// even number of items
	func() {
		b := []byte{'a', 'b', 'c', 'd'}
		res := ReverseBytes(b[:])
		assert.Equal(t, b, res)
		assert.Equal(t, []byte{'d', 'c', 'b', 'a'}, b)
	}()
}
