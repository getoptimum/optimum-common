package rlnc

import (
	"errors"
	"fmt"
	"slices"
)

// ErrInvalidBasis is returned when a basis is not in reduced row-echelon form.
var ErrInvalidBasis = errors.New("invalid basis")

// Reduce removes the directions of basis from row in place. basis must be in
// reduced row-echelon form, as built by Insert. Rows are zero before their
// pivot, so only the suffix from each pivot needs field arithmetic.
//
// pivots[i] is the leading column of basis[i]. If pivots does not have one
// entry per row, leading columns are found by scanning each row.
func Reduce(basis [][]byte, pivots []int, row []byte) {
	for i, existing := range basis {
		pivot := pivotAt(basis, pivots, i)
		if factor := row[pivot]; factor != 0 {
			XorScaled(row[pivot:], existing[pivot:], factor)
		}
	}
}

// Insert adds row, which must already be reduced against basis and nonzero at
// pivot, its leading column. The row is copied and scaled so its pivot is one,
// then eliminated from the existing rows and inserted in pivot order, preserving
// reduced row-echelon form without sorting or reducing the whole basis again.
// Existing rows are modified in place, so the caller must own them.
//
// It returns the updated basis and pivots. As in Reduce, pivots that do not
// have one entry per row are rebuilt from basis.
func Insert(basis [][]byte, pivots []int, row []byte, pivot int) (newBasis [][]byte, newPivots []int) {
	inserted := make([]byte, len(row))

	copy(inserted, row)
	Scale(inserted[pivot:], GFInverse(inserted[pivot]))

	for _, existing := range basis {
		if factor := existing[pivot]; factor != 0 {
			XorScaled(existing[pivot:], inserted[pivot:], factor)
		}
	}

	if len(pivots) != len(basis) {
		pivots = make([]int, len(basis))
		for i, existing := range basis {
			pivots[i] = LeadingColumn(existing)
		}
	}

	index, _ := slices.BinarySearch(pivots, pivot)

	return slices.Insert(basis, index, inserted), slices.Insert(pivots, index, pivot)
}

// ValidateRREF checks that basis is in reduced row-echelon form and that every
// row has the given width.
func ValidateRREF(basis [][]byte, width int) error {
	previous := -1

	for index, row := range basis {
		if len(row) != width {
			return fmt.Errorf("basis width: %w", ErrInvalidBasis)
		}

		pivot := LeadingColumn(row)
		if pivot <= previous || pivot < 0 || row[pivot] != 1 {
			return fmt.Errorf("basis pivot: %w", ErrInvalidBasis)
		}

		// Increasing leading columns imply zeros below every pivot. Only the
		// earlier rows need checking for zeros above it, so no pivot buffer is needed.
		for _, earlier := range basis[:index] {
			if earlier[pivot] != 0 {
				return fmt.Errorf("basis not reduced: %w", ErrInvalidBasis)
			}
		}
		previous = pivot
	}

	return nil
}

func pivotAt(basis [][]byte, pivots []int, index int) int {
	if len(pivots) == len(basis) {
		return pivots[index]
	}

	return LeadingColumn(basis[index])
}
