package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInSlice(t *testing.T) {
	assert.True(t, InSlice(1, []any{3, 2, 1}), "Int in slice")
	assert.True(t, InSlice(1, []any{3.04, "2", "1", 1}), "Int in mixed slice")
	assert.False(t, InSlice("1", []any{3.04, "2", 1}), "Int not in mixed slice")
}

func TestSliceIntToAny(t *testing.T) {
	converted := SliceIntToAny([]int{1, 2, 3})
	assert.Len(t, converted, 3, "Length is unchanged")
	assert.Equal(t, []any{1, 2, 3}, converted, "Arrays should be equal")
}

func TestSliceStrToAny(t *testing.T) {
	converted := SliceStrToAny([]string{"1", "2", "3"})
	assert.Len(t, converted, 3, "Length is unchanged")
	assert.Equal(t, []any{"1", "2", "3"}, converted, "Arrays should be equal")
}
