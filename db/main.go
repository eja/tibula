// Package db provides functions for database operations, including opening and closing connections,
// running queries, and retrieving results. It supports SQLite and MySQL database engines.
//
// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"

	"github.com/eja/tibula/log"
)

const tag = "[db]"

type TypeSession struct {
	Handler      *sql.DB
	Engine       string
	ConnectionId int64
}

// Session start a new database session
func Session() TypeSession {
	return TypeSession{}
}

// Open initializes a connection to the specified database using the provided parameters.
func (session *TypeSession) Open(engine string, database string, username string, password string, host string, port int) (err error) {

	if database == "" {
		return errors.New("database name/file is mandatory")
	}

	switch engine {
	case "sqlite":
		session.Handler, err = sqliteOpen(database)
	case "mysql":
		if username != "" && password != "" {
			session.Handler, err = mysqlOpen(database, username, password, host, port)
		} else {
			return errors.New("username/password missing")
		}
	default:
		return errors.New("unsupported database engine")
	}

	if err == nil {
		session.Engine = engine
		session.ConnectionId += 1
		log.Debug(tag, "open", session.Engine)
	}

	return
}

// Close closes the open database connection.
func (session *TypeSession) Close() error {
	if session.Handler != nil {
		log.Debug(tag, "close", session.Engine)
		return session.Handler.Close()
	}
	return errors.New("no database connection to close")
}

// Run executes a query with optional parameters and returns a TypeRun containing information about the execution.
func (session *TypeSession) Run(query string, args ...interface{}) (result TypeRun, err error) {
	switch session.Engine {
	case "sqlite":
		result, err = session.sqliteRun(query, args...)
	case "mysql":
		result, err = session.mysqlRun(query, args...)
	default:
		err = errors.New("engine not found")
	}

	if err != nil {
		log.Error(tag, err, query, args, err)
	} else {
		log.Trace(tag, query, args)
	}
	return
}

// Value executes a query with optional parameters and returns a single result as a string.
func (session *TypeSession) Value(query string, args ...interface{}) (result string, err error) {
	switch session.Engine {
	case "sqlite":
		result, err = session.sqliteValue(query, args...)
	case "mysql":
		result, err = session.mysqlValue(query, args...)
	default:
		err = errors.New("engine not found")
	}
	if err == sql.ErrNoRows {
		err = nil
	}

	if err != nil {
		log.Error(tag, err, query, args)
	} else {
		log.Trace(tag, query, args)
	}
	return
}

// Row executes a query with optional parameters and returns a single row of results as a TypeRow.
func (session *TypeSession) Row(query string, args ...interface{}) (result TypeRow, err error) {
	switch session.Engine {
	case "sqlite":
		result, err = session.sqliteRow(query, args...)
	case "mysql":
		result, err = session.mysqlRow(query, args...)
	default:
		err = errors.New("engine not found")
	}
	if err == sql.ErrNoRows {
		err = nil
	}

	if err != nil {
		log.Error(tag, err, query, args)
	} else {
		log.Trace(tag, query, args)
	}
	return
}

// Rows executes a query with optional parameters and returns multiple rows of results as a TypeRows.
func (session *TypeSession) Rows(query string, args ...interface{}) (result TypeRows, err error) {
	switch session.Engine {
	case "sqlite":
		result, err = session.sqliteRows(query, args...)
	case "mysql":
		result, err = session.mysqlRows(query, args...)
	default:
		err = errors.New("engine not found")
	}
	if err == sql.ErrNoRows {
		err = nil
	}

	if err != nil {
		log.Error(tag, err, query, args)
	} else {
		log.Trace(tag, query, args)
	}
	return
}

// Cols executes a query with optional parameters and returns the column names of the result set.
func (session *TypeSession) Cols(query string, args ...interface{}) ([]string, error) {
	switch session.Engine {
	case "sqlite":
		return session.sqliteCols(query, args...)
	case "mysql":
		return session.mysqlCols(query, args...)
	default:
		return nil, errors.New("engine not found")
	}
}
