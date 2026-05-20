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

func TestGenerateBulkInsertSQL(t *testing.T) {
	t.Parallel()
	type row struct{ ID int }
	q, params := sql.GenerateBulkInsertSQL("items", sql.PQParamPlaceholder, []row{{1}, {2}}, func(r row) map[string]any {
		return map[string]any{"id": r.ID}
	})
	require.Contains(t, q, "INSERT INTO items")
	require.Contains(t, q, "VALUES")
	require.Len(t, params, 2)
}

func TestGenerateBulkUpsertSQL_Update(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateBulkUpsertSQL(
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
	require.Equal(t,
		"INSERT INTO heads (chain_id,slot,block_root) VALUES ($1,$2,$3),($4,$5,$6) ON CONFLICT (chain_id, slot) DO UPDATE SET block_root = EXCLUDED.block_root",
		q,
	)
	require.Len(t, params, 6)
}

func TestGenerateBulkUpsertSQL_DoNothing(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateBulkUpsertSQL(
		"duties",
		[]string{"chain_id", "slot", "proposer_index"},
		[]string{"chain_id", "slot"},
		nil,
		1,
		func(i int) []any {
			return []any{"hoodi", uint64(10), uint64(42)}
		},
	)
	require.Equal(t,
		"INSERT INTO duties (chain_id,slot,proposer_index) VALUES ($1,$2,$3) ON CONFLICT (chain_id, slot) DO NOTHING",
		q,
	)
	require.Len(t, params, 3)
}

func TestGenerateBulkUpsertSQL_Empty(t *testing.T) {
	t.Parallel()
	q, params := sql.GenerateBulkUpsertSQL("t", nil, nil, nil, 0, nil)
	require.Empty(t, q)
	require.Nil(t, params)
}
