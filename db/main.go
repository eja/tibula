// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

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

func Session() TypeSession {
	return TypeSession{}
}

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
			if err == nil {
				_, err = session.Handler.Exec("SET sql_mode = 'PIPES_AS_CONCAT'")
			}
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

func (session *TypeSession) Close() error {
	if session.Handler != nil {
		log.Debug(tag, "close", session.Engine)
		return session.Handler.Close()
	}
	return errors.New("no database connection to close")
}

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
