// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

// New creates a new entry in the specified module table with the given ownerId and moduleId.
func (session *TypeSession) New(ownerId int64, moduleId int64) (int64, error) {
	moduleName := session.ModuleGetNameById(moduleId)

	check, err := session.TableExists(moduleName)
	if err != nil {
		return 0, err
	}
	if !check {
		return 0, errors.New("table does not exist")
	}

	query := fmt.Sprintf("INSERT INTO %s (ejaOwner, ejaLog) VALUES (?,?)", moduleName)
	run, err := session.Run(query, ownerId, session.Now())
	if err != nil {
		return 0, err
	}
	return run.LastId, nil
}

// Get retrieves a row from the specified module table based on ownerId, moduleId, and ejaId.
func (session *TypeSession) Get(ownerId int64, moduleId int64, ejaId int64) (TypeRow, error) {
	moduleName := session.ModuleGetNameById(moduleId)

	check, err := session.TableExists(moduleName)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, errors.New("table does not exist")
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE ejaId=? AND ejaOwner IN (%s)", moduleName, session.OwnersCsv(ownerId, moduleId))
	return session.Row(query, ejaId)
}

// Put updates a specific field in a row of the specified module table based on ownerId, moduleId, ejaId, fieldName, and fieldValue.
func (session *TypeSession) Put(ownerId int64, moduleId int64, ejaId int64, fieldName string, fieldValue interface{}) error {
	moduleName := session.ModuleGetNameById(moduleId)

	check, err := session.TableExists(moduleName)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table not found")
	}

	check, err = session.FieldExists(moduleName, fieldName)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("field not found")
	}

	query := fmt.Sprintf("UPDATE %s SET %s=? WHERE ejaId=? AND ejaOwner IN (%s)", moduleName, fieldName, session.OwnersCsv(ownerId, moduleId))
	_, err = session.Run(query, fieldValue, ejaId)
	if err != nil {
		return err
	}
	return nil
}

// Del deletes a row from the specified module table based on ownerId, moduleId, and ejaId.
func (session *TypeSession) Del(ownerId int64, moduleId, ejaId int64) error {
	owners := session.Owners(ownerId, moduleId)
	csv := session.NumbersToCsv(owners)
	moduleName := session.ModuleGetNameById(moduleId)

	if moduleName == "ejaModules" {
		ejaModulesOwnersCsv := session.OwnersCsv(ownerId, moduleId)
		tableName, err := session.Value("SELECT name FROM ejaModules WHERE ejaId=? AND ejaOwner IN ("+ejaModulesOwnersCsv+")", ejaId)
		if err == nil && tableName != "" {
			session.TableDel(tableName)
			session.Run("DELETE FROM ejaFields WHERE ejaModuleId=?", ejaId)
			session.Run("DELETE FROM ejaPermissions WHERE ejaModuleId=?", ejaId)
			session.Run("DELETE FROM ejaTranslations WHERE ejaModuleId=?", ejaId)
			session.Run("DELETE FROM ejaModuleLinks WHERE dstModuleId=?", ejaId)
		}
	}

	// Delete the entry from the module table
	query := fmt.Sprintf("DELETE FROM %s WHERE ejaId=? AND ejaOwner IN (%s)", moduleName, csv)
	if _, err := session.Run(query, ejaId); err != nil {
		return err
	}

	// Delete related entries from 'ejaLinks' table
	query = fmt.Sprintf("DELETE FROM ejaLinks WHERE (dstModuleId=? AND dstFieldId=?) OR (srcModuleId=? AND srcFieldId=?) AND ejaOwner IN (%s)", csv)
	if _, err := session.Run(query, moduleId, ejaId, moduleId, ejaId); err != nil {
		return err
	}

	return nil
}
