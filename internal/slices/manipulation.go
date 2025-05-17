// Contains slice index manipulation functions like duplicating or subsets
package slices

// Last returns the final element of a given slice
func Last[T any](slice []T) T {
	return slice[len(slice)-1]
}

// AllExcept returns every element in slice except for the last N
func AllExcept[T any](slice []T, n int) []T {
	// If longer than slice, return an empty one
	if n >= len(slice) {
		return []T{}
	}
	// Otherwise do normal indexing
	return slice[:len(slice)-n]
}

// CopySlice returns a new slice with the same elements as the given slice
func CopySlice[T interface{}](slice []T) []T {
	n := make([]T, len(slice))
	copy(n, slice)

	return n
}

// CopySliceWith returns a new slice with the same elements as the given slice, plus the additional elements
func CopySliceWith[T interface{}](slice []T, x ...T) []T {
	return append(CopySlice(slice), x...)
}
