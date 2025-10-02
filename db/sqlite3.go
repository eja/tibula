// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

func sqliteOpen(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}

func (session *TypeSession) sqliteRun(query string, args ...interface{}) (TypeRun, error) {
	result, err := session.Handler.Exec(query, args...)
	if err != nil {
		return TypeRun{}, err
	}
	lastId, _ := result.LastInsertId()
	changes, _ := result.RowsAffected()
	return TypeRun{Changes: changes, LastId: lastId}, nil
}

func (session *TypeSession) sqliteValue(query string, args ...interface{}) (result string, err error) {
	var nullResult sql.NullString
	err = session.Handler.QueryRow(query, args...).Scan(&nullResult)
	if err != nil {
		return
	}
	if nullResult.Valid {
		result = nullResult.String
	} else {
		result = ""
	}

	return
}

func (session *TypeSession) sqliteRow(query string, args ...interface{}) (TypeRow, error) {
	var result TypeRow
	rows, err := session.sqliteRows(query, args...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result = row
	}
	return result, nil
}

func (session *TypeSession) sqliteRows(query string, args ...interface{}) (TypeRows, error) {
	rows, err := session.Handler.Query(query, args...)
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

func (session *TypeSession) sqliteCols(query string, args ...interface{}) ([]string, error) {
	rows, err := session.Handler.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows.Columns()
}

func (session *TypeSession) sqliteTableExists(name string) (bool, error) {
	if err := sqliteTableNameIsValid(name); err != nil {
		return false, err
	}
	if _, err := session.sqliteValue("SELECT name FROM sqlite_master WHERE type='table' AND name=?", name); err != nil {
		if _, err := session.sqliteValue("SELECT name FROM sqlite_temp_master WHERE type='table' AND name=?", name); err != nil {
			return false, nil
		}
	}
	return true, nil
}

func (session *TypeSession) sqliteFieldExists(tableName, fieldName string) (bool, error) {
	if err := sqliteTableNameIsValid(tableName); err != nil {
		return false, err
	}

	if err := sqliteFieldNameIsValid(fieldName); err != nil {
		return false, err
	}

	rows, err := session.sqliteRows(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
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
