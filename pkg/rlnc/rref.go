package rlnc

import (
	"errors"
	"fmt"
	"slices"
)

// ErrInvalidBasis is returned when a basis is not in reduced row-echelon form.
var ErrInvalidBasis = errors.New("invalid basis")

// Reduce removes the directions of basis from row in place. basis must be in
// row-echelon form with each pivot equal to one, as built by Insert in either
// mode. Rows are zero before their pivot, so only the suffix from each pivot
// needs field arithmetic.
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
// then inserted in pivot order. It returns the updated basis and pivots, which
// must have one entry per row.
//
// With clearColumn, the new pivot's column is also eliminated from the existing
// rows, so the basis stays in reduced row-echelon form: the same coding space
// always gives the same rows, whatever order rows arrive in. Without it, the
// existing rows are left as they are, which is cheaper but order dependent.
// clearColumn modifies existing rows in place, so the caller must own them.
func Insert(basis [][]byte, pivots []int, row []byte, pivot int, clearColumn bool) (newBasis [][]byte, newPivots []int) {
	inserted := make([]byte, len(row))

	copy(inserted, row)
	Scale(inserted[pivot:], Inverse(inserted[pivot]))

	if clearColumn {
		for _, existing := range basis {
			if factor := existing[pivot]; factor != 0 {
				XorScaled(existing[pivot:], inserted[pivot:], factor)
			}
		}
	}

	index, _ := slices.BinarySearch(pivots, pivot)

	return slices.Insert(basis, index, inserted), slices.Insert(pivots, index, pivot)
}

// ValidateRREF checks that basis is in reduced row-echelon form and that every
// row has the given width. A basis built by Insert without clearColumn fails
// this check.
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
