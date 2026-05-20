package sql

import (
	"fmt"
	"sort"
	"strings"
)

const PQParamPlaceholder = "$"

func GenerateInsertSQL(tableName string, fieldsValuesMapping map[string]any) (sql string, params []any) {
	keys := sortedMapKeys(fieldsValuesMapping)
	fields := make([]string, 0, len(keys))
	placeholders := make([]string, 0, len(keys))
	params = make([]any, 0, len(keys))
	counter := 1
	for _, k := range keys {
		v := fieldsValuesMapping[k]
		params = append(params, v)
		fields = append(fields, k)
		placeholders = append(placeholders, fmt.Sprintf("$%d", counter))
		counter++
	}
	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(fields, ", "), strings.Join(placeholders, ", "))
	return sql, params
}

func GenerateUpsertSQL(tableName string, fieldsValuesMapping map[string]any, conflictColumns []string) (sql string, params []any) {
	sql, params = GenerateInsertSQL(tableName, fieldsValuesMapping)

	conflictSet := make(map[string]struct{}, len(conflictColumns))
	for _, c := range conflictColumns {
		conflictSet[c] = struct{}{}
	}

	setClauses := make([]string, 0, len(fieldsValuesMapping))
	for _, k := range sortedMapKeys(fieldsValuesMapping) {
		if _, isConflict := conflictSet[k]; !isConflict {
			setClauses = append(setClauses, fmt.Sprintf("%s = EXCLUDED.%s", k, k))
		}
	}

	conflict := strings.Join(conflictColumns, ", ")
	if len(setClauses) == 0 {
		sql += fmt.Sprintf(" ON CONFLICT (%s) DO NOTHING", conflict)
	} else {
		sql += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s", conflict, strings.Join(setClauses, ", "))
	}

	return sql, params
}

// GenerateBulkInsertSQL returns ("", nil, nil) when entityList is empty.
func GenerateBulkInsertSQL[T any](
	tableName string,
	paramPlaceholder string,
	entityList []T,
	entityProcessor func(entity T) map[string]any,
) (sql string, params []any, err error) {
	if len(entityList) == 0 {
		return "", nil, nil
	}
	columns := columnsFromProcessor(entityList[0], entityProcessor)
	placeholders, rowParams, err := buildBulkValues(paramPlaceholder, columns, len(entityList), func(i int) []any {
		m := entityProcessor(entityList[i])
		row := make([]any, len(columns))
		for j, col := range columns {
			row[j] = m[col]
		}
		return row
	})
	if err != nil {
		return "", nil, err
	}
	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	return sql, rowParams, nil
}

// GenerateBulkUpsertSQL with empty updateColumns emits ON CONFLICT … DO NOTHING.
// Returns an error if any row from rowValues has fewer values than len(columns).
func GenerateBulkUpsertSQL(
	tableName string,
	columns, conflictColumns, updateColumns []string,
	rowCount int,
	rowValues func(i int) []any,
) (query string, args []any, err error) {
	if rowCount == 0 {
		return "", nil, nil
	}
	placeholders, rowParams, err := buildBulkValues(PQParamPlaceholder, columns, rowCount, rowValues)
	if err != nil {
		return "", nil, err
	}
	conflict := strings.Join(conflictColumns, ", ")
	if len(updateColumns) == 0 {
		query = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO NOTHING",
			tableName, strings.Join(columns, ","), strings.Join(placeholders, ","), conflict,
		)
		args = rowParams
		return query, args, nil
	}
	updateSet := make([]string, len(updateColumns))
	for i, c := range updateColumns {
		updateSet[i] = fmt.Sprintf("%s = EXCLUDED.%s", c, c)
	}
	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
		conflict,
		strings.Join(updateSet, ", "),
	)
	args = rowParams
	return query, args, nil
}

func sortedMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func columnsFromProcessor[T any](sample T, entityProcessor func(entity T) map[string]any) []string {
	m := entityProcessor(sample)
	keys := sortedMapKeys(m)
	return keys
}

func buildBulkValues(paramPlaceholder string, columns []string, rowCount int, rowValues func(i int) []any) (placeholders []string, params []any, err error) {
	colCount := len(columns)
	placeholders = make([]string, 0, rowCount)
	params = make([]any, 0, rowCount*colCount)
	counter := 1
	for i := 0; i < rowCount; i++ {
		row := rowValues(i)
		if len(row) < colCount {
			return nil, nil, fmt.Errorf("pkg/sql: row %d has %d values, expected %d", i, len(row), colCount)
		}
		ph := make([]string, colCount)
		for j := range columns {
			params = append(params, row[j])
			ph[j] = fmt.Sprintf("%s%d", paramPlaceholder, counter)
			counter++
		}
		placeholders = append(placeholders, "("+strings.Join(ph, ",")+")")
	}
	return placeholders, params, nil
}
