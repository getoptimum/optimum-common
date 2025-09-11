package sliceutil

func ConvertSlice[T, U any](src []T, converter func(T) U) []U {
	result := make([]U, len(src)) // preallocate with exact length
	for i, v := range src {
		result[i] = converter(v)
	}
	return result
}
