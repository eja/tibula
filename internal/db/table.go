// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

// TableExists checks if a table with the specified name exists in the database.
func TableExists(name string) (bool, error) {
	switch DbEngine {
	case "sqlite":
		return sqliteTableExists(name)
	case "mysql":
		return mysqlTableExists(name)
	default:
		return false, errors.New("engine not found")
	}
}

// TableNameIsValid checks if a table name is valid according to the current database engine's conventions.
func TableNameIsValid(name string) error {
	switch DbEngine {
	case "sqlite":
		return sqliteTableNameIsValid(name)
	case "mysql":
		return mysqlTableNameIsValid(name)
	default:
		return errors.New("engine not found")
	}
}

// TableAdd creates a new table with the specified name in the database.
// If the table already exists, it does nothing.
// The optional 'tmp' parameter specifies whether the table is temporary.
func TableAdd(name string, tmp ...bool) error {
	check, err := TableExists(name)
	if err != nil {
		return err
	}
	if !check {
		temporary := ""
		if len(tmp) > 0 {
			temporary = "TEMPORARY"
		}
		switch DbEngine {
		case "sqlite":
			if _, err := Run(fmt.Sprintf("CREATE %s TABLE %s (ejaId INTEGER PRIMARY KEY, ejaOwner INTEGER, ejaLog DATETIME)", temporary, name)); err != nil {
				return err
			}
		case "mysql":
			if _, err := Run(fmt.Sprintf("CREATE %s TABLE %s (ejaId INTEGER AUTO_INCREMENT PRIMARY KEY, ejaOwner INTEGER, ejaLog DATETIME)", temporary, name)); err != nil {
				return err
			}
		default:
			return errors.New("engine not found")
		}
	}
	return nil
}

// TableDel deletes the table with the specified name from the database.
// If the table does not exist, it returns an error.
func TableDel(name string) error {
	check, err := TableExists(name)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table does not exists")
	}
	if _, err := Run(fmt.Sprintf("DROP TABLE %s", name)); err != nil {
		return err
	}
	return nil
}

// TableGetAllById retrieves a row from the specified table based on the ejaId field.
func TableGetAllById(tableName string, ejaId int64) TypeRow {
	row, _ := Row("SELECT * FROM "+tableName+" WHERE ejaId=?", ejaId)
	return row
}
