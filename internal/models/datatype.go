package models

import (
	"database/sql"
	"fmt"
	"strings"
)

type GoType int

const (
	TypeNull GoType = iota
	TypeString
	TypeInt
	TypeFloat
	TypeBool
	TypeTime
	TypeBytes
	TypeJSON
	TypeArray
	TypeOther
)

func DetecGoType(collType string) GoType {
	lower := strings.ToLower(collType)
	switch {
	case contains(lower, "int"), contains(lower, "serial"), lower == "bool":
		return TypeInt
	case contains(lower, "float"), contains(lower, "double"), contains(lower, "decimal"), contains(lower, "numeric"), contains(lower, "real"):
		return TypeFloat
	case contains(lower, "char"), contains(lower, "text"), contains(lower, "varchar"), contains(lower, "clob"):
		return TypeString
	case lower == "bool", lower == "boolean":
		return TypeBool
	case contains(lower, "time"), contains(lower, "date"), lower == "timestamp":
		return TypeTime
	case contains(lower, "blob"), contains(lower, "binary"), contains(lower, "bytea"):
		return TypeBytes
	case contains(lower, "json"):
		return TypeJSON
	default:
		return TypeOther
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

const maxCellLen = 1024

func truncateCell(s string) string {
	if len(s) > maxCellLen {
		return s[:maxCellLen] + "..."
	}
	return s
}

func FormatCellValue(val any) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case []byte:
		return truncateCell(string(v))
	case fmt.Stringer:
		return truncateCell(v.String())
	default:
		s := fmt.Sprintf("%v", v)
		if len(s) > maxCellLen {
			s = s[:maxCellLen] + "..."
		}
		return s
	}
}

func ScanRow(columns []*sql.ColumnType, row *sql.Row) ([]any, error) {
	vals := make([]any, len(columns))
	for i := range vals {
		vals[i] = new(any)
	}
	if err := row.Scan(vals...); err != nil {
		return nil, err
	}
	result := make([]any, len(vals))
	for i, v := range vals {
		result[i] = *(v.(*any))
	}
	return result, nil
}

func ScanRows(columns []*sql.ColumnType, rows *sql.Rows) ([][]any, error) {
	var results [][]any
	for rows.Next() {
		vals := make([]any, len(columns))
		for i := range vals {
			vals[i] = new(any)
		}
		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}
		row := make([]any, len(vals))
		for i, v := range vals {
			row[i] = *(v.(*any))
		}
		results = append(results, row)
	}
	return results, rows.Err()
}
