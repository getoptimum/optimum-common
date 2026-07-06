package sql_test

import (
	"strings"
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
	require.Contains(t, qs(q), "on conflict (id) do update set")
	require.Contains(t, qs(q), "name = excluded.name")
	require.Len(t, params, 2)
}

func TestGenerateUpsert_AllConflictKeys_DoNothing(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateUpsertSQL("items", map[string]any{"id": 1}, []string{"id"})
	require.Contains(t, qs(q), "on conflict (id) do nothing")
	require.NotContains(t, qs(q), "do update set")
	require.Len(t, params, 1)
}

func TestGenerateUpsert_NoConflictSuffix(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateUpsertSQL("items", map[string]any{"id": 1}, nil)
	require.NotContains(t, qs(q), "on conflict")
	require.Len(t, params, 1)
}

func TestGenerateBulkInsertSQL(t *testing.T) {
	t.Parallel()
	type row struct{ ID int }
	q, params := sql.GenerateBulkInsertSQL("items", sql.PQParamPlaceholder, []row{{1}, {2}}, func(r row) map[string]any {
		return map[string]any{"id": r.ID}
	})
	require.Contains(t, q, "INSERT INTO items")
	require.Len(t, params, 2)
}

func TestGenerateBulkInsertSQL_Empty(t *testing.T) {
	t.Parallel()
	type row struct{ ID int }
	q, params := sql.GenerateBulkInsertSQL("items", sql.PQParamPlaceholder, []row(nil), func(r row) map[string]any {
		return map[string]any{"id": r.ID}
	})
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

func TestGenerateBulkUpsertSQL_EmptyRowCount(t *testing.T) {
	t.Parallel()
	q, params, err := sql.GenerateBulkUpsertSQL("t", nil, []string{"a"}, nil, 0, nil)
	require.NoError(t, err)
	require.Empty(t, q)
	require.Nil(t, params)
}

func TestGenerateBulkUpsertSQL_NoConflictErrors(t *testing.T) {
	t.Parallel()
	_, _, err := sql.GenerateBulkUpsertSQL("t", []string{"a"}, nil, nil, 1, func(i int) []any { return []any{1} })
	require.Error(t, err)
}

func TestGenerateBulkUpsertSQL_RowTooShort(t *testing.T) {
	t.Parallel()
	_, _, err := sql.GenerateBulkUpsertSQL(
		"t",
		[]string{"a", "b"},
		[]string{"a"},
		[]string{"b"},
		1,
		func(i int) []any { return []any{1} },
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "row 0")
}

func TestGenerateInsertSQL_DeterministicColumnOrder(t *testing.T) {
	t.Parallel()
	m := map[string]any{"zulu": 1, "alpha": 2, "mike": 3}
	wantSQL := "INSERT INTO items (alpha, mike, zulu) VALUES ($1, $2, $3)"
	wantParams := []any{2, 3, 1} // values in the same sorted order as the columns
	for range 50 {
		q, params := sql.GenerateInsertSQL("items", m)
		require.Equal(t, wantSQL, q)
		require.Equal(t, wantParams, params)
	}
}

// TestGenerateUpsertSQL_DeterministicSetOrder guards the DO UPDATE SET clause order.
func TestGenerateUpsertSQL_DeterministicSetOrder(t *testing.T) {
	t.Parallel()
	m := map[string]any{"id": 1, "zulu": "z", "alpha": "a", "mike": "m"}
	want := "INSERT INTO items (alpha, id, mike, zulu) VALUES ($1, $2, $3, $4) " +
		"ON CONFLICT (id) DO UPDATE SET alpha = EXCLUDED.alpha, mike = EXCLUDED.mike, zulu = EXCLUDED.zulu"
	for range 50 {
		q, _ := sql.GenerateUpsertSQL("items", m, []string{"id"})
		require.Equal(t, want, q)
	}
}

func qs(s string) string { return strings.ToLower(s) }
