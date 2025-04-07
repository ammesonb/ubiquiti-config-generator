package slices

// Returns true if elem is in arr
func InSlice(elem any, arr []any) bool {
	for _, x := range arr {
		if x == elem {
			return true
		}
	}

	return false
}

// Returns a new slice with the same elements as arr, but typed as any
func SliceStrToAny(arr []string) []any {
	anyArr := make([]any, len(arr))

	for i := range arr {
		anyArr[i] = arr[i]
	}

	return anyArr
}

// Returns a new slice with the same elements as arr, but typed as any
func SliceIntToAny(arr []int) []any {
	anyArr := make([]any, len(arr))

	for i := range arr {
		anyArr[i] = arr[i]
	}

	return anyArr
}
