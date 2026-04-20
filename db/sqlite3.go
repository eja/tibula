// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	_ "modernc.org/sqlite"
)

var (
	sqliteValidTableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,127}$`)
	sqliteValidFieldNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,127}$`)
)

func sqliteOpen(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)", path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (session *TypeSession) sqliteRun(query string, args ...any) (TypeRun, error) {
	result, err := session.Handler.Exec(query, args...)
	if err != nil {
		return TypeRun{}, err
	}
	lastId, _ := result.LastInsertId()
	changes, _ := result.RowsAffected()
	return TypeRun{Changes: changes, LastId: lastId}, nil
}

func (session *TypeSession) sqliteValue(query string, args ...any) (string, error) {
	var nullResult sql.NullString
	err := session.Handler.QueryRow(query, args...).Scan(&nullResult)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if nullResult.Valid {
		return nullResult.String, nil
	}
	return "", nil
}

func (session *TypeSession) sqliteRow(query string, args ...any) (TypeRow, error) {
	rows, err := session.Handler.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	if err := rows.Scan(scanArgs...); err != nil {
		return nil, err
	}

	result := make(TypeRow)
	for i, col := range values {
		if col == nil {
			result[columns[i]] = ""
		} else {
			result[columns[i]] = string(col)
		}
	}

	return result, nil
}

func (session *TypeSession) sqliteRows(query string, args ...any) (TypeRows, error) {
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
	scanArgs := make([]any, len(values))
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
			if col == nil {
				row[columns[i]] = ""
			} else {
				row[columns[i]] = string(col)
			}
		}
		result = append(result, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (session *TypeSession) sqliteCols(query string, args ...any) ([]string, error) {
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
	query := `
		SELECT count(*) FROM (
			SELECT name FROM sqlite_master WHERE type='table' AND name=? 
			UNION ALL 
			SELECT name FROM sqlite_temp_master WHERE type='table' AND name=?
		)
	`
	var count int
	err := session.Handler.QueryRow(query, name, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (session *TypeSession) sqliteFieldExists(tableName, fieldName string) (bool, error) {
	if err := sqliteTableNameIsValid(tableName); err != nil {
		return false, err
	}
	if err := sqliteFieldNameIsValid(fieldName); err != nil {
		return false, err
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?", tableName)

	var count int
	err := session.Handler.QueryRow(query, fieldName).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func sqliteTableNameIsValid(name string) error {
	if !sqliteValidTableNameRegex.MatchString(name) {
		return errors.New("table name is not valid")
	}
	return nil
}

func sqliteFieldNameIsValid(name string) error {
	if !sqliteValidFieldNameRegex.MatchString(name) {
		return errors.New("field name is not valid")
	}
	return nil
}

func (session *TypeSession) sqliteFtsAdd(tableName, columnName string) error {
	if err := sqliteTableNameIsValid(tableName); err != nil {
		return err
	}
	if err := sqliteFieldNameIsValid(columnName); err != nil {
		return err
	}

	ftsTableName := fmt.Sprintf("ejaFTS_%s_%s", tableName, columnName)

	createFts := fmt.Sprintf(
		"CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(%s, content='%s', content_rowid='rowid')",
		ftsTableName, columnName, tableName,
	)
	if _, err := session.Handler.Exec(createFts); err != nil {
		return err
	}

	ai := fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_ai AFTER INSERT ON %s BEGIN
			INSERT INTO %s(rowid, %s) VALUES (new.rowid, new.%s);
		END;`, ftsTableName, tableName, ftsTableName, columnName, columnName)

	ad := fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_ad AFTER DELETE ON %s BEGIN
			INSERT INTO %s(%s, rowid, %s) VALUES('delete', old.rowid, old.%s);
		END;`, ftsTableName, tableName, ftsTableName, ftsTableName, columnName, columnName)

	au := fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_au AFTER UPDATE ON %s BEGIN
			INSERT INTO %s(%s, rowid, %s) VALUES('delete', old.rowid, old.%s);
			INSERT INTO %s(rowid, %s) VALUES(new.rowid, new.%s);
		END;`, ftsTableName, tableName, ftsTableName, ftsTableName, columnName, columnName, ftsTableName, columnName, columnName)

	session.Handler.Exec(ai)
	session.Handler.Exec(ad)
	session.Handler.Exec(au)

	syncSql := fmt.Sprintf("INSERT INTO %s(rowid, %s) SELECT rowid, %s FROM %s",
		ftsTableName, columnName, columnName, tableName)
	session.Handler.Exec(syncSql)

	return nil
}
