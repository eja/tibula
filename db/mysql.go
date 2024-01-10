// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

// mysqlOpen establishes a MySQL database connection using the provided parameters.
func mysqlOpen(database string, username string, password string, host string, port int) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
	return sql.Open("mysql", connectionString)
}

// mysqlRun executes a SQL query on the MySQL database and returns information about the execution.
func mysqlRun(query string, args ...interface{}) (TypeRun, error) {
	result, err := DbHandler.Exec(query, args...)
	if err != nil {
		return TypeRun{}, err
	}
	lastId, _ := result.LastInsertId()
	changes, _ := result.RowsAffected()
	return TypeRun{Changes: changes, LastId: lastId}, nil
}

// mysqlValue executes a SQL query on the MySQL database and returns a single result as a string.
func mysqlValue(query string, args ...interface{}) (result string, err error) {
	row := DbHandler.QueryRow(query, args...)
	err = row.Scan(&result)
	if err != nil {
		return
	}
	return
}

// mysqlRow executes a SQL query on the MySQL database and returns a single row of results as a TypeRow.
func mysqlRow(query string, args ...interface{}) (TypeRow, error) {
	var result TypeRow
	rows, err := mysqlRows(query, args...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result = row
		break
	}
	return result, nil
}

// mysqlRows executes a SQL query on the MySQL database and returns multiple rows of results as a TypeRows.
func mysqlRows(query string, args ...interface{}) (TypeRows, error) {
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

// mysqlCols executes a SQL query on the MySQL database and returns the column names of the result set.
func mysqlCols(query string, args ...interface{}) ([]string, error) {
	rows, err := DbHandler.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows.Columns()
}

// mysqlTableExists checks if a table with the given name exists in the MySQL database.
func mysqlTableExists(name string) (bool, error) {
	if err := mysqlTableNameIsValid(name); err != nil {
		return false, err
	}

	mysqlRun("SET @dbName = DATABASE()")
	mysqlRun("CALL sys.table_exists(@dbName,'" + name + "',@tableExists)")
	exists, err := mysqlValue("SELECT @tableExists")
	if err != nil {
		return false, err
	}
	return exists != "", nil
}

// mysqlFieldExists checks if a field with the given name exists in the specified table of the MySQL database.
func mysqlFieldExists(tableName, fieldName string) (bool, error) {
	if err := mysqlTableNameIsValid(tableName); err != nil {
		return false, err
	}

	if err := mysqlFieldNameIsValid(fieldName); err != nil {
		return false, err
	}

	rows, err := mysqlRows(fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", tableName, fieldName))
	if err != nil {
		return false, err
	}
	if len(rows) > 0 {
		return true, nil
	}
	return false, nil
}

// mysqlTableNameIsValid checks if a table name is valid based on MySQL naming conventions.
func mysqlTableNameIsValid(name string) error {
	check, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`, name)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table name is not valid")
	}
	return nil
}

// mysqlFieldNameIsValid checks if a field name is valid based on MySQL naming conventions.
func mysqlFieldNameIsValid(name string) error {
	check, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`, name)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("field name is not valid")
	}
	return nil
}
