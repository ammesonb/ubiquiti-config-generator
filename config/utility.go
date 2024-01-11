package config

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
