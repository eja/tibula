// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

// sqliteOpen opens a connection to an SQLite database at the specified path.
func sqliteOpen(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}

// sqliteRun executes a SQL query with optional arguments and returns information about the execution.
func sqliteRun(query string, args ...interface{}) (TypeRun, error) {
	result, err := DbHandler.Exec(query, args...)
	if err != nil {
		return TypeRun{}, err
	}
	lastId, _ := result.LastInsertId()
	changes, _ := result.RowsAffected()
	return TypeRun{Changes: changes, LastId: lastId}, nil
}

// sqliteValue executes a SQL query with optional arguments and returns a single string result.
func sqliteValue(query string, args ...interface{}) (result string, err error) {
	err = DbHandler.QueryRow(query, args...).Scan(&result)
	if err != nil {
		return
	}
	return
}

// sqliteRow executes a SQL query with optional arguments and returns a single row of results.
func sqliteRow(query string, args ...interface{}) (TypeRow, error) {
	var result TypeRow
	rows, err := sqliteRows(query, args...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result = row
	}
	return result, nil
}

// sqliteRows executes a SQL query with optional arguments and returns multiple rows of results.
func sqliteRows(query string, args ...interface{}) (TypeRows, error) {
	rows, err := DbHandler.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result TypeRows
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		row := make(TypeRow)
		for i, col := range values {
			row[columns[i]] = string(col)
		}

		result = append(result, row)
	}

	return result, nil
}

// sqliteCols executes a SQL query with optional arguments and returns the column names.
func sqliteCols(query string, args ...interface{}) ([]string, error) {
	rows, err := DbHandler.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows.Columns()
}

// sqliteTableExists checks if a table with the specified name exists in the database.
func sqliteTableExists(name string) (bool, error) {
	if err := sqliteTableNameIsValid(name); err != nil {
		return false, err
	}
	if _, err := sqliteValue("SELECT name FROM sqlite_master WHERE type='table' AND name=?", name); err != nil {
		if _, err := sqliteValue("SELECT name FROM sqlite_temp_master WHERE type='table' AND name=?", name); err != nil {
			return false, nil
		}
	}
	return true, nil
}

// sqliteFieldExists checks if a field with the specified name exists in the specified table.
func sqliteFieldExists(tableName, fieldName string) (bool, error) {
	if err := sqliteTableNameIsValid(tableName); err != nil {
		return false, err
	}

	if err := sqliteFieldNameIsValid(fieldName); err != nil {
		return false, err
	}

	rows, err := sqliteRows(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}

	for _, row := range rows {
		if row["name"] == fieldName {
			return true, nil
		}
	}

	return false, nil
}

// sqliteTableNameIsValid checks if a table name is valid according to SQLite naming conventions.
func sqliteTableNameIsValid(name string) error {
	check, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{0,127}$`, name)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table name is not valid")
	}
	return nil
}

// sqliteFieldNameIsValid checks if a field name is valid according to SQLite naming conventions.
func sqliteFieldNameIsValid(name string) error {
	check, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{0,127}$`, name)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("field name is not valid")
	}
	return nil
}
