package slices

import "slices"

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

// UniqueComparable returns a new slice containing only unique elements from the
// input. Uniqueness is determined by ==. Order is preserved (first occurrence kept).
func UniqueComparable[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// ContainsInSlice checks if value exists in slice.
// Returns true if found, false otherwise.
func ContainsInSlice[T comparable](slice []T, value T) bool {
	return slices.Contains(slice, value)
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
// If chunkSize <= 0, nil is returned.
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(slice)+chunkSize-1)/chunkSize)
	for i := 0; i < len(slice); i += chunkSize {
		end := min(i+chunkSize, len(slice))
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// AppendBounded appends v while capping at limit elements, shifting out the oldest when full.
// If limit <= 0, s is unchanged.
func AppendBounded[T any](s []T, v T, limit int) []T {
	if limit <= 0 {
		return s
	}
	if len(s) < limit {
		return append(s, v)
	}
	copy(s, s[1:])
	s[len(s)-1] = v
	return s
}

// KeepLast returns a new slice with the last up to limit elements, or nil if limit <= 0.
func KeepLast[T any](s []T, limit int) []T {
	if limit <= 0 {
		return nil
	}
	if len(s) <= limit {
		return append([]T(nil), s...)
	}
	return append([]T(nil), s[len(s)-limit:]...)
}

// ConcatKeepLast flattens chunks, then KeepLast(flat, limit).
func ConcatKeepLast[T any](chunks [][]T, limit int) []T {
	if limit <= 0 {
		return nil
	}
	var n int
	for _, c := range chunks {
		n += len(c)
	}
	flat := make([]T, 0, n)
	for _, c := range chunks {
		flat = append(flat, c...)
	}
	return KeepLast(flat, limit)
}
