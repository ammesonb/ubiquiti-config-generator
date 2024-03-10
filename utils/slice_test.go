package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopySlice(t *testing.T) {
	slice := []string{"a", "b", "c"}
	copied := CopySlice(slice)
	assert.Equal(t, slice, copied)
}

func TestCopySliceWith(t *testing.T) {
	slice := []string{"a", "b", "c"}
	copied := CopySliceWith(slice, "d")
	assert.Equal(t, []string{"a", "b", "c", "d"}, copied)
}

func TestLast(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.Equal(t, "c", Last(slice))
}

func TestAllExcept(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.Equal(t, []string{"a", "b"}, AllExcept(slice, 1))
	assert.Equal(t, []string{}, AllExcept(slice, 5))
}
