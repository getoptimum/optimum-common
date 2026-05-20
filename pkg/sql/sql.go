package sql

import (
	"fmt"
	"strings"
)

const PQParamPlaceholder = "$"

func GenerateInsertSQL(tableName string, fieldsValuesMapping map[string]any) (sql string, params []any) {
	fields := make([]string, 0, len(fieldsValuesMapping))
	placeholders := make([]string, 0, len(fieldsValuesMapping))
	params = make([]any, 0, len(fieldsValuesMapping))
	counter := 1
	for k, v := range fieldsValuesMapping {
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
	for k := range fieldsValuesMapping {
		if _, isConflict := conflictSet[k]; !isConflict {
			setClauses = append(setClauses, fmt.Sprintf("%s = EXCLUDED.%s", k, k))
		}
	}

	sql += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s",
		strings.Join(conflictColumns, ", "),
		strings.Join(setClauses, ", "))

	return sql, params
}

// GenerateBulkInsertSQL panics if entityList is empty (columns are taken from the first row).
func GenerateBulkInsertSQL[T any](
	tableName string,
	paramPlaceholder string,
	entityList []T,
	entityProcessor func(entity T) map[string]any,
) (sql string, params []any) {
	columns := columnsFromProcessor(entityList[0], entityProcessor)
	placeholders, params := buildBulkValues(paramPlaceholder, columns, len(entityList), func(i int) []any {
		m := entityProcessor(entityList[i])
		row := make([]any, len(columns))
		for j, col := range columns {
			row[j] = m[col]
		}
		return row
	})
	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	return sql, params
}

// GenerateBulkUpsertSQL with empty updateColumns emits ON CONFLICT … DO NOTHING.
func GenerateBulkUpsertSQL(
	tableName string,
	columns, conflictColumns, updateColumns []string,
	rowCount int,
	rowValues func(i int) []any,
) (string, []any) {
	if rowCount == 0 {
		return "", nil
	}
	placeholders, params := buildBulkValues(PQParamPlaceholder, columns, rowCount, rowValues)
	conflict := strings.Join(conflictColumns, ", ")
	if len(updateColumns) == 0 {
		return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO NOTHING",
			tableName, strings.Join(columns, ","), strings.Join(placeholders, ","), conflict,
		), params
	}
	updateSet := make([]string, len(updateColumns))
	for i, c := range updateColumns {
		updateSet[i] = fmt.Sprintf("%s = EXCLUDED.%s", c, c)
	}
	q := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
		conflict,
		strings.Join(updateSet, ", "),
	)
	return q, params
}

func columnsFromProcessor[T any](sample T, entityProcessor func(entity T) map[string]any) []string {
	columns := make([]string, 0, 16)
	for k := range entityProcessor(sample) {
		columns = append(columns, k)
	}
	return columns
}

func buildBulkValues(paramPlaceholder string, columns []string, rowCount int, rowValues func(i int) []any) (placeholders []string, params []any) {
	colCount := len(columns)
	placeholders = make([]string, 0, rowCount)
	params = make([]any, 0, rowCount*colCount)
	counter := 1
	for i := 0; i < rowCount; i++ {
		row := rowValues(i)
		ph := make([]string, colCount)
		for j := range columns {
			params = append(params, row[j])
			ph[j] = fmt.Sprintf("%s%d", paramPlaceholder, counter)
			counter++
		}
		placeholders = append(placeholders, "("+strings.Join(ph, ",")+")")
	}
	return placeholders, params
}
