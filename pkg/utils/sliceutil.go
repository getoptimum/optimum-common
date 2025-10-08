package utils

import "fmt"

func MapSlice[T, U any](src []T, converter func(T) U) []U {
	result := make([]U, len(src)) // preallocate with exact length
	for i, v := range src {
		result[i] = converter(v)
	}
	return result
}

func UniqueSlice[T fmt.Stringer](slice []T) []T {
	seen := make(map[string]struct{})
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v.String()]; !ok {
			seen[v.String()] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
func ContainsInSlice[T comparable](slice []T, value T) bool {
	for i := range slice {
		if slice[i] == value {
			return true
		}
	}
	return false
}
