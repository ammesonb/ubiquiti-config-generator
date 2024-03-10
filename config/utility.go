package config

func InSlice(elem any, arr []any) bool {
	for _, x := range arr {
		if x == elem {
			return true
		}
	}

	return false
}

func SliceStrToAny(arr []string) []any {
	anyArr := make([]any, len(arr))

	for i := range arr {
		anyArr[i] = arr[i]
	}

	return anyArr
}

func SliceIntToAny(arr []int) []any {
	anyArr := make([]any, len(arr))

	for i := range arr {
		anyArr[i] = arr[i]
	}

	return anyArr
}
