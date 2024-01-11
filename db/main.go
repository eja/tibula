// Package db provides functions for database operations, including opening and closing connections,
// running queries, and retrieving results. It supports SQLite and MySQL database engines.
//
// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
)

// DbHandler is the database handler representing the open database connection.
var DbHandler *sql.DB

// DbEngine holds the current database engine in use (e.g., "sqlite", "mysql").
var DbEngine string

// DbConnectionId holds the current database connection
var DbConnectionId int64

// Open initializes a connection to the specified database using the provided parameters.
func Open(engine string, database string, username string, password string, host string, port int) error {
	var err error

	if database == "" {
		return errors.New("database name/file is mandatory")
	}

	switch engine {
	case "sqlite":
		DbHandler, err = sqliteOpen(database)
	case "mysql":
		if username != "" && password != "" {
			DbHandler, err = mysqlOpen(database, username, password, host, port)
		} else {
			return errors.New("username/password missing")
		}
	default:
		return errors.New("unsupported database engine")
	}

	if err == nil {
		DbEngine = engine
		DbConnectionId += 1
		Debug("db open", DbEngine)
	}
	return err
}

// Close closes the open database connection.
func Close() error {
	if DbHandler != nil {
		Debug("db close", DbEngine)
		return DbHandler.Close()
	}
	return errors.New("no database connection to close")
}

// Run executes a query with optional parameters and returns a TypeRun containing information about the execution.
func Run(query string, args ...interface{}) (result TypeRun, err error) {
	switch DbEngine {
	case "sqlite":
		result, err = sqliteRun(query, args...)
	case "mysql":
		result, err = mysqlRun(query, args...)
	default:
		err = errors.New("engine not found")
	}

	if err != nil {
		Error(err, query, args, err)
	} else {
		Trace(query, args)
	}
	return
}

// Value executes a query with optional parameters and returns a single result as a string.
func Value(query string, args ...interface{}) (result string, err error) {
	switch DbEngine {
	case "sqlite":
		result, err = sqliteValue(query, args...)
	case "mysql":
		result, err = mysqlValue(query, args...)
	default:
		err = errors.New("engine not found")
	}
	if err == sql.ErrNoRows {
		err = nil
	}

	if err != nil {
		Error(err, query, args)
	} else {
		Trace(query, args)
	}
	return
}

// Row executes a query with optional parameters and returns a single row of results as a TypeRow.
func Row(query string, args ...interface{}) (result TypeRow, err error) {
	switch DbEngine {
	case "sqlite":
		result, err = sqliteRow(query, args...)
	case "mysql":
		result, err = mysqlRow(query, args...)
	default:
		err = errors.New("engine not found")
	}
	if err == sql.ErrNoRows {
		err = nil
	}

	if err != nil {
		Error(err, query, args)
	} else {
		Trace(query, args)
	}
	return
}

// Rows executes a query with optional parameters and returns multiple rows of results as a TypeRows.
func Rows(query string, args ...interface{}) (result TypeRows, err error) {
	switch DbEngine {
	case "sqlite":
		result, err = sqliteRows(query, args...)
	case "mysql":
		result, err = mysqlRows(query, args...)
	default:
		err = errors.New("engine not found")
	}
	if err == sql.ErrNoRows {
		err = nil
	}

	if err != nil {
		Error(err, query, args)
	} else {
		Trace(query, args)
	}
	return
}

// Cols executes a query with optional parameters and returns the column names of the result set.
func Cols(query string, args ...interface{}) ([]string, error) {
	switch DbEngine {
	case "sqlite":
		return sqliteCols(query, args...)
	case "mysql":
		return mysqlCols(query, args...)
	default:
		return nil, errors.New("engine not found")
	}
}
