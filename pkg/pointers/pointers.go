package pointers

// ToPointer returns a pointer to the provided value.
//
//go:fix inline
func ToPointer[T any](value T) *T {
	return new(value)
}

// FromPointer returns the value pointed to by value, or the zero value of T if value is nil.
func FromPointer[T any](value *T) T {
	var result T
	if value == nil {
		return result
	}
	return *value
}
