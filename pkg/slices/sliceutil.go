package slices

import "fmt"

// MapSlice applies a converter function to each element of src and returns a new slice.
// The result slice has the same length as src.
func MapSlice[T, U any](src []T, converter func(T) U) []U {
	result := make([]U, len(src)) // preallocate with exact length
	for i, v := range src {
		result[i] = converter(v)
	}
	return result
}

// UniqueSlice returns a new slice containing only unique elements from the input.
// Uniqueness is determined by the String() method of each element.
// The order of elements is preserved (first occurrence is kept).
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

// ContainsInSlice checks if value exists in slice.
// Returns true if found, false otherwise.
func ContainsInSlice[T comparable](slice []T, value T) bool {
	for i := range slice {
		if slice[i] == value {
			return true
		}
	}
	return false
}

// ExcludeFromSlice returns a new slice with all values from excludeValues removed from slice
// not optimized for performance
func ExcludeFromSlice[T comparable](slice, excludeValues []T) []T {
	result := make([]T, 0, len(slice))
	for _, entry := range slice {
		if !ContainsInSlice(excludeValues, entry) {
			result = append(result, entry)
		}
	}
	return result
}

// ChunkSlice splits a slice into chunks of the specified size.
// The last chunk may be smaller than chunkSize if the slice length is not divisible by chunkSize.
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	chunks := make([][]T, 0, (len(slice)+chunkSize-1)/chunkSize)
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
