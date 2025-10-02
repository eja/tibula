// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

func (session *TypeSession) TableExists(name string) (bool, error) {
	switch session.Engine {
	case "sqlite":
		return session.sqliteTableExists(name)
	case "mysql":
		return session.mysqlTableExists(name)
	default:
		return false, errors.New("engine not found")
	}
}

func (session *TypeSession) TableNameIsValid(name string) error {
	switch session.Engine {
	case "sqlite":
		return sqliteTableNameIsValid(name)
	case "mysql":
		return mysqlTableNameIsValid(name)
	default:
		return errors.New("engine not found")
	}
}

func (session *TypeSession) TableAdd(name string, tmp ...bool) error {
	check, err := session.TableExists(name)
	if err != nil {
		return err
	}
	if !check {
		temporary := ""
		if len(tmp) > 0 {
			temporary = "TEMPORARY"
		}
		switch session.Engine {
		case "sqlite":
			if _, err := session.Run(fmt.Sprintf("CREATE %s TABLE %s (ejaId INTEGER PRIMARY KEY, ejaOwner INTEGER, ejaLog DATETIME)", temporary, name)); err != nil {
				return err
			}
		case "mysql":
			if _, err := session.Run(fmt.Sprintf("CREATE %s TABLE %s (ejaId INTEGER AUTO_INCREMENT PRIMARY KEY, ejaOwner INTEGER, ejaLog DATETIME)", temporary, name)); err != nil {
				return err
			}
		default:
			return errors.New("engine not found")
		}
	}
	return nil
}

func (session *TypeSession) TableDel(name string) error {
	check, err := session.TableExists(name)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table does not exists")
	}
	if _, err := session.Run(fmt.Sprintf("DROP TABLE %s", name)); err != nil {
		return err
	}
	return nil
}

func (session *TypeSession) TableGetAllById(tableName string, ejaId int64) TypeRow {
	row, _ := session.Row("SELECT * FROM "+tableName+" WHERE ejaId=?", ejaId)
	return row
}
