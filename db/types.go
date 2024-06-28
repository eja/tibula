// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"strconv"
)

// TypeRun represents the result of a database operation, including the number of changes and the last inserted ID.
type TypeRun struct {
	Changes int64
	LastId  int64
}

// TypeRow is a map representing a single row of data in the database.
type TypeRow map[string]string

// TypeRows is a slice of maps representing multiple rows of data in the database.
type TypeRows []map[string]string

// TypeSelect represents a key-value pair used for dropdowns or selection lists.
type TypeSelect struct {
	Key   string
	Value string
}

// String converts a value to a string.
// It handles various types, including string, int, int64, float32, float64, and []uint8.
func (session *TypeSession) String(nameValue interface{}) string {
	switch v := nameValue.(type) {
	case string:
		return v
	case int, int64, float32, float64:
		return strconv.FormatInt(session.Number(nameValue), 10)
	case []uint8:
		return string(v)
	default:
		return ""
	}
}

// Number converts a value to an int64.
// It handles various numeric types and attempts to parse a string as an int64.
// If parsing fails, it returns 0.
func (session *TypeSession) Number(nameValue interface{}) int64 {
	switch v := nameValue.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		value := session.String(v)
		num, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return num
		}
	}
	return 0
}

// Float converts a value to a float64.
// It handles various numeric types and attempts to parse a string as a float64.
// If parsing fails, it returns 0.0.
func (session *TypeSession) Float(nameValue interface{}) float64 {
	switch v := nameValue.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return num
		}
	}
	return 0.0
}
