package utils

// Last returns the final element of a given slice
func Last[T interface{}](slice []T) T {
	return slice[len(slice)-1]
}

// AllExcept returns every element in slice except for the last N
func AllExcept[T interface{}](slice []T, n int) []T {
	// If longer than slice, return an empty one
	if n >= len(slice) {
		return []T{}
	}
	// Otherwise do normal indexing
	return slice[:len(slice)-n]
}

func CopySlice[T interface{}](slice []T) []T {
	n := make([]T, len(slice))
	for i := range slice {
		n[i] = slice[i]
	}

	return n
}

func CopySliceWith[T interface{}](slice []T, x ...T) []T {
	return append(CopySlice(slice), x...)
}
