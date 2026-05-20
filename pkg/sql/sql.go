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

// GenerateUpsertSQL appends ON CONFLICT only when conflictColumns is non-empty.
func GenerateUpsertSQL(tableName string, fieldsValuesMapping map[string]any, conflictColumns []string) (sql string, params []any) {
	sql, params = GenerateInsertSQL(tableName, fieldsValuesMapping)
	if len(conflictColumns) == 0 {
		return sql, params
	}

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

	conflict := strings.Join(conflictColumns, ", ")
	if len(setClauses) == 0 {
		sql += fmt.Sprintf(" ON CONFLICT (%s) DO NOTHING", conflict)
	} else {
		sql += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s", conflict, strings.Join(setClauses, ", "))
	}
	return sql, params
}

func GenerateBulkInsertSQL[T any](
	tableName string,
	paramPlaceholder string,
	entityList []T,
	entityProcessor func(entity T) map[string]any,
) (sql string, params []any) {
	if len(entityList) == 0 {
		return "", nil
	}
	columns := columnsFromMap(entityProcessor(entityList[0]))
	placeholders, rowParams := buildBulkInsertRows(paramPlaceholder, columns, len(entityList), func(i int) []any {
		m := entityProcessor(entityList[i])
		row := make([]any, len(columns))
		for j, col := range columns {
			row[j] = m[col]
		}
		return row
	})
	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, strings.Join(columns, ","), strings.Join(placeholders, ","))
	return sql, rowParams
}

func buildBulkInsertRows(paramPlaceholder string, columns []string, rowCount int, rowValues func(i int) []any) (placeholders []string, params []any) {
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

// GenerateBulkUpsertSQL requires non-empty conflictColumns. Empty updateColumns => DO NOTHING on that conflict.
func GenerateBulkUpsertSQL(
	tableName string,
	columns, conflictColumns, updateColumns []string,
	rowCount int,
	rowValues func(i int) []any,
) (query string, args []any, err error) {
	if rowCount == 0 {
		return "", nil, nil
	}
	if len(conflictColumns) == 0 {
		return "", nil, fmt.Errorf("pkg/sql: conflict columns required")
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

func columnsFromMap(m map[string]any) []string {
	columns := make([]string, 0, len(m))
	for k := range m {
		columns = append(columns, k)
	}
	return columns
}

func buildBulkValues(paramPlaceholder string, columns []string, rowCount int, rowValues func(i int) []any) (placeholders []string, params []any, err error) {
	colCount := len(columns)
	placeholders = make([]string, 0, rowCount)
	params = make([]any, 0, rowCount*colCount)
	counter := 1
	for i := 0; i < rowCount; i++ {
		row := rowValues(i)
		if len(row) < colCount {
			return nil, nil, fmt.Errorf("pkg/sql: row %d: need %d values, got %d", i, colCount, len(row))
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
