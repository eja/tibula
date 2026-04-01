// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
	"log/slog"
)

const (
	SESSION_EXPIRE = 10000 //>2 <6 hours
)

var tag = slog.String("module", "tibula.db")

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
		slog.Debug("open", tag, "engine", session.Engine)
	}

	return
}

func (session *TypeSession) Close() error {
	if session.Handler != nil {
		slog.Debug("close", tag, "engine", session.Engine)
		return session.Handler.Close()
	}
	return errors.New("no database connection to close")
}

func (session *TypeSession) Run(query string, args ...any) (result TypeRun, err error) {
	switch session.Engine {
	case "sqlite":
		result, err = session.sqliteRun(query, args...)
	case "mysql":
		result, err = session.mysqlRun(query, args...)
	default:
		err = errors.New("engine not found")
	}

	if err != nil {
		slog.Error("query run error", tag, "query", query, "args", args, "error", err)
	} else {
		slog.Debug("query run", tag, "query", query, "args", args)
	}
	return
}

func (session *TypeSession) Value(query string, args ...any) (result string, err error) {
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
		slog.Error("query value error", tag, "query", query, "args", args, "error", err)
	} else {
		slog.Debug("query value", tag, "query", query, "args", args)
	}
	return
}

func (session *TypeSession) Row(query string, args ...any) (result TypeRow, err error) {
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
		slog.Error("query row error", tag, "query", query, "args", args, "error", err)
	} else {
		slog.Debug("query row", tag, "query", query, "args", args)
	}
	return
}

func (session *TypeSession) Rows(query string, args ...any) (result TypeRows, err error) {
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
		slog.Error("query rows error", tag, "query", query, "args", args, "error", err)
	} else {
		slog.Debug("query rows", tag, "query", query, "args", args)
	}
	return
}

func (session *TypeSession) Cols(query string, args ...any) ([]string, error) {
	switch session.Engine {
	case "sqlite":
		return session.sqliteCols(query, args...)
	case "mysql":
		return session.mysqlCols(query, args...)
	default:
		return nil, errors.New("engine not found")
	}
}
