package sql_test

import (
	"testing"

	"github.com/getoptimum/optimum-common/pkg/sql"
	"github.com/stretchr/testify/require"
)

func TestGenerateInsertSQL(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateInsertSQL("items", map[string]any{"id": 1, "name": "a"})
	require.Contains(t, q, "INSERT INTO items")
	require.Contains(t, q, "VALUES")
	require.Len(t, params, 2)
}

func TestGenerateUpsertSQL(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateUpsertSQL("items", map[string]any{"id": 1, "name": "a"}, []string{"id"})
	require.Contains(t, q, "ON CONFLICT (id) DO UPDATE SET")
	require.Contains(t, q, "name = EXCLUDED.name")
	require.Len(t, params, 2)
}

func TestGenerateUpsertAllConflictKeys_DoNothing(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateUpsertSQL("items", map[string]any{"id": 1}, []string{"id"})
	require.Contains(t, q, "ON CONFLICT (id) DO NOTHING")
	require.NotContains(t, q, "DO UPDATE SET")
	require.Len(t, params, 1)
}

func TestGenerateBulkInsertSQL(t *testing.T) {
	t.Parallel()
	type row struct{ ID int }
	q, params, err := sql.GenerateBulkInsertSQL("items", sql.PQParamPlaceholder, []row{{1}, {2}}, func(r row) map[string]any {
		return map[string]any{"id": r.ID}
	})
	require.NoError(t, err)
	require.Contains(t, q, "INSERT INTO items")
	require.Contains(t, q, "VALUES")
	require.Len(t, params, 2)
}

func TestGenerateBulkInsertSQL_Empty(t *testing.T) {
	t.Parallel()
	type row struct{ ID int }
	q, params, err := sql.GenerateBulkInsertSQL("items", sql.PQParamPlaceholder, []row(nil), func(r row) map[string]any {
		return map[string]any{"id": r.ID}
	})
	require.NoError(t, err)
	require.Empty(t, q)
	require.Nil(t, params)
}

func TestGenerateBulkUpsertSQL_Update(t *testing.T) {
	t.Parallel()
	q, params, err := sql.GenerateBulkUpsertSQL(
		"heads",
		[]string{"chain_id", "slot", "block_root"},
		[]string{"chain_id", "slot"},
		[]string{"block_root"},
		2,
		func(i int) []any {
			if i == 0 {
				return []any{"hoodi", uint64(1), "0xa"}
			}
			return []any{"hoodi", uint64(2), "0xb"}
		},
	)
	require.NoError(t, err)
	require.Equal(t,
		"INSERT INTO heads (chain_id,slot,block_root) VALUES ($1,$2,$3),($4,$5,$6) ON CONFLICT (chain_id, slot) DO UPDATE SET block_root = EXCLUDED.block_root",
		q,
	)
	require.Len(t, params, 6)
}

func TestGenerateBulkUpsertSQL_DoNothing(t *testing.T) {
	t.Parallel()
	q, params, err := sql.GenerateBulkUpsertSQL(
		"duties",
		[]string{"chain_id", "slot", "proposer_index"},
		[]string{"chain_id", "slot"},
		nil,
		1,
		func(i int) []any {
			return []any{"hoodi", uint64(10), uint64(42)}
		},
	)
	require.NoError(t, err)
	require.Equal(t,
		"INSERT INTO duties (chain_id,slot,proposer_index) VALUES ($1,$2,$3) ON CONFLICT (chain_id, slot) DO NOTHING",
		q,
	)
	require.Len(t, params, 3)
}

func TestGenerateBulkUpsertSQL_Empty(t *testing.T) {
	t.Parallel()
	q, params, err := sql.GenerateBulkUpsertSQL("t", nil, nil, nil, 0, nil)
	require.NoError(t, err)
	require.Empty(t, q)
	require.Nil(t, params)
}

func TestGenerateBulkUpsertSQL_RowTooShort(t *testing.T) {
	t.Parallel()
	_, _, err := sql.GenerateBulkUpsertSQL(
		"t",
		[]string{"a", "b"},
		[]string{"a"},
		[]string{"b"},
		1,
		func(i int) []any {
			return []any{1}
		},
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "row 0")
}

func TestDeterministicOrdering_InsertAndUpsert(t *testing.T) {
	t.Parallel()
	m := map[string]any{"zebra": 1, "alpha": 2, "mid": 3}
	wantInsert := "INSERT INTO t (alpha, mid, zebra) VALUES ($1, $2, $3)"
	for range 5 {
		q, params := sql.GenerateInsertSQL("t", m)
		require.Equal(t, wantInsert, q)
		require.Equal(t, []any{2, 3, 1}, params)
	}
	for range 5 {
		q, _ := sql.GenerateUpsertSQL("t", m, []string{"alpha"})
		require.Contains(t, q, "(alpha, mid, zebra)")
		require.Contains(t, q, "mid = EXCLUDED.mid")
		require.Contains(t, q, "zebra = EXCLUDED.zebra")
	}
}
