package rlnc_test

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getoptimum/optimum-common/pkg/rlnc"
)

// addRows builds a basis from rows with Reduce and Insert, skipping rows that
// add no new direction.
func addRows(rows [][]byte, clearColumn bool) (basis [][]byte, pivots []int) {
	for _, row := range rows {
		candidate := slices.Clone(row)
		rlnc.Reduce(basis, pivots, candidate)

		if pivot := rlnc.LeadingColumn(candidate); pivot >= 0 {
			basis, pivots = rlnc.Insert(basis, pivots, candidate, pivot, clearColumn)
		}
	}

	return basis, pivots
}

// referenceBasis uses column-first elimination independently of the
// incremental Reduce and Insert.
func referenceBasis(rows [][]byte, k int) [][]byte {
	basis := make([][]byte, len(rows))
	for i, row := range rows {
		basis[i] = slices.Clone(row)
	}

	rank := 0
	for column := 0; column < k && rank < len(basis); column++ {
		pivot := rank
		for pivot < len(basis) && basis[pivot][column] == 0 {
			pivot++
		}
		if pivot == len(basis) {
			continue
		}

		basis[rank], basis[pivot] = basis[pivot], basis[rank]
		rlnc.Scale(basis[rank], rlnc.GFInverse(basis[rank][column]))
		for i, row := range basis {
			if i != rank {
				rlnc.XorScaled(row, basis[rank], row[column])
			}
		}
		rank++
	}

	return basis[:rank]
}

func randomRows(rng *rand.Rand, n, k int, dependent bool) [][]byte {
	rows := make([][]byte, n)
	for i := range rows {
		rows[i] = make([]byte, k)
		for column := range k {
			rows[i][column] = byte(rng.Uint32())
		}
	}

	// Make the second half scaled copies of the first, so half the rows are
	// redundant.
	if dependent {
		for i := n / 2; i < n; i++ {
			rows[i] = slices.Clone(rows[i-n/2])
			rlnc.Scale(rows[i], byte(i+1))
		}
	}

	return rows
}

func TestInsertClearColumnMatchesElimination(t *testing.T) {
	rng := rand.New(rand.NewPCG(12, 34))

	for _, k := range []int{3, 16, 64} {
		for _, dependent := range []bool{false, true} {
			t.Run(fmt.Sprintf("K=%d/dependent=%t", k, dependent), func(t *testing.T) {
				rows := randomRows(rng, k, k, dependent)

				basis, pivots := addRows(rows, true)
				want := referenceBasis(rows, k)

				require.Equal(t, want, basis)
				require.NoError(t, rlnc.ValidateRREF(basis, k))
				for i, row := range basis {
					require.Equal(t, rlnc.LeadingColumn(row), pivots[i])
				}
			})
		}
	}
}

func TestInsertWithoutClearColumnKeepsRank(t *testing.T) {
	rng := rand.New(rand.NewPCG(56, 78))

	for _, k := range []int{3, 16, 64} {
		for _, dependent := range []bool{false, true} {
			t.Run(fmt.Sprintf("K=%d/dependent=%t", k, dependent), func(t *testing.T) {
				rows := randomRows(rng, k, k, dependent)

				basis, pivots := addRows(rows, false)
				require.Len(t, basis, len(referenceBasis(rows, k)))

				for i, row := range basis {
					require.Equal(t, rlnc.LeadingColumn(row), pivots[i])
					require.Equal(t, byte(1), row[pivots[i]])
				}
				require.True(t, slices.IsSorted(pivots))

				// Every input row lies in the span of the basis.
				for _, row := range rows {
					candidate := slices.Clone(row)
					rlnc.Reduce(basis, pivots, candidate)
					require.Equal(t, -1, rlnc.LeadingColumn(candidate))
				}
			})
		}
	}
}

func TestInsertModes(t *testing.T) {
	rows := [][]byte{{1, 3, 7}, {0, 1, 5}}

	// Row 0 keeps the 3 above row 1's pivot.
	basis, _ := addRows(rows, false)
	require.Equal(t, [][]byte{{1, 3, 7}, {0, 1, 5}}, basis)
	require.ErrorIs(t, rlnc.ValidateRREF(basis, 3), rlnc.ErrInvalidBasis)

	// Row 0 has 3 times row 1 added to it, clearing column 1.
	basis, _ = addRows(rows, true)
	require.Equal(t, [][]byte{{1, 0, 7 ^ rlnc.GFMultiply(3, 5)}, {0, 1, 5}}, basis)
	require.NoError(t, rlnc.ValidateRREF(basis, 3))
}

func TestReduceWithoutPivots(t *testing.T) {
	basis, pivots := addRows([][]byte{{1, 2, 0}, {0, 0, 1}}, true)

	withPivots := []byte{3, 6, 9}
	rlnc.Reduce(basis, pivots, withPivots)

	withoutPivots := []byte{3, 6, 9}
	rlnc.Reduce(basis, nil, withoutPivots)

	require.Equal(t, withPivots, withoutPivots)
	require.Equal(t, []byte{0, 0, 0}, withoutPivots)
}

func TestValidateRREF(t *testing.T) {
	testCases := []struct {
		name      string
		basis     [][]byte
		width     int
		expectErr bool
	}{
		{name: "empty", width: 3},
		{name: "identity", basis: [][]byte{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, width: 3},
		{name: "reduced", basis: [][]byte{{1, 4, 0}, {0, 0, 1}}, width: 3},
		{name: "wrong width", basis: [][]byte{{1, 0}}, width: 3, expectErr: true},
		{name: "zero row", basis: [][]byte{{0, 0, 0}}, width: 3, expectErr: true},
		{name: "pivot not one", basis: [][]byte{{2, 0, 0}}, width: 3, expectErr: true},
		{name: "pivots out of order", basis: [][]byte{{0, 1, 0}, {1, 0, 0}}, width: 3, expectErr: true},
		{name: "repeated pivot", basis: [][]byte{{1, 0, 0}, {1, 1, 0}}, width: 3, expectErr: true},
		{name: "not reduced", basis: [][]byte{{1, 3, 0}, {0, 1, 0}}, width: 3, expectErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := rlnc.ValidateRREF(tc.basis, tc.width)

			if tc.expectErr {
				require.ErrorIs(t, err, rlnc.ErrInvalidBasis)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
