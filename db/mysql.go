// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

func mysqlOpen(database string, username string, password string, host string, port int) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
	return sql.Open("mysql", connectionString)
}

func (session *TypeSession) mysqlRun(query string, args ...interface{}) (TypeRun, error) {
	result, err := session.Handler.Exec(query, args...)
	if err != nil {
		return TypeRun{}, err
	}
	lastId, _ := result.LastInsertId()
	changes, _ := result.RowsAffected()
	return TypeRun{Changes: changes, LastId: lastId}, nil
}

func (session *TypeSession) mysqlValue(query string, args ...interface{}) (result string, err error) {
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

func (session *TypeSession) mysqlRow(query string, args ...interface{}) (TypeRow, error) {
	var result TypeRow
	rows, err := session.mysqlRows(query, args...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result = row
		break
	}
	return result, nil
}

func (session *TypeSession) mysqlRows(query string, args ...interface{}) (TypeRows, error) {
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

func (session *TypeSession) mysqlCols(query string, args ...interface{}) ([]string, error) {
	rows, err := session.Handler.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows.Columns()
}

func (session *TypeSession) mysqlTableExists(name string) (bool, error) {
	if err := mysqlTableNameIsValid(name); err != nil {
		return false, err
	}

	//we need this approach to be able to check also for temporary tables
	session.mysqlRun("SET @dbName = DATABASE()")
	session.mysqlRun("CALL sys.table_exists(@dbName,'" + name + "',@tableExists)")
	exists, err := session.mysqlValue("SELECT @tableExists")
	if err != nil {
		return false, err
	}

	return exists != "", nil
}

func (session *TypeSession) mysqlFieldExists(tableName, fieldName string) (bool, error) {
	if err := mysqlTableNameIsValid(tableName); err != nil {
		return false, err
	}

	if err := mysqlFieldNameIsValid(fieldName); err != nil {
		return false, err
	}

	rows, err := session.mysqlRows(fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", tableName, fieldName))
	if err != nil {
		return false, err
	}
	if len(rows) > 0 {
		return true, nil
	}
	return false, nil
}

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
