// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

// New creates a new entry in the specified module table with the given ownerId and moduleId.
func New(ownerId int64, moduleId int64) (int64, error) {
	moduleName := ModuleGetNameById(moduleId)

	check, err := TableExists(moduleName)
	if err != nil {
		return 0, err
	}
	if !check {
		return 0, errors.New("table does not exist")
	}

	query := fmt.Sprintf("INSERT INTO %s (ejaOwner, ejaLog) VALUES (?,?)", moduleName)
	run, err := Run(query, ownerId, Now())
	if err != nil {
		return 0, err
	}
	return run.LastId, nil
}

// Get retrieves a row from the specified module table based on ownerId, moduleId, and ejaId.
func Get(ownerId int64, moduleId int64, ejaId int64) (TypeRow, error) {
	moduleName := ModuleGetNameById(moduleId)

	check, err := TableExists(moduleName)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, errors.New("table does not exist")
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE ejaId=? AND ejaOwner IN (%s)", moduleName, OwnersCsv(ownerId, moduleId))
	return Row(query, ejaId)
}

// Put updates a specific field in a row of the specified module table based on ownerId, moduleId, ejaId, fieldName, and fieldValue.
func Put(ownerId int64, moduleId int64, ejaId int64, fieldName string, fieldValue string) error {
	moduleName := ModuleGetNameById(moduleId)

	check, err := TableExists(moduleName)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table not found")
	}

	check, err = FieldExists(moduleName, fieldName)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("field not found")
	}

	query := fmt.Sprintf("UPDATE %s SET %s=? WHERE ejaId=? AND ejaOwner IN (%s)", moduleName, fieldName, OwnersCsv(ownerId, moduleId))
	_, err = Run(query, fieldValue, ejaId)
	if err != nil {
		return err
	}
	return nil
}

// Del deletes a row from the specified module table based on ownerId, moduleId, and ejaId.
func Del(ownerId int64, moduleId, ejaId int64) error {
	owners := Owners(ownerId, moduleId)
	csv := NumbersToCsv(owners)
	moduleName := ModuleGetNameById(moduleId)

	if moduleName == "ejaModules" {
		ejaModulesOwnersCsv := OwnersCsv(ownerId, moduleId)
		tableName, err := Value("SELECT name FROM ejaModules WHERE ejaId=? AND ejaOwner IN ("+ejaModulesOwnersCsv+")", ejaId)
		if err == nil && tableName != "" {
			TableDel(tableName)
			Run("DELETE FROM ejaFields WHERE ejaModuleId=?", ejaId)
			Run("DELETE FROM ejaPermissions WHERE ejaModuleId=?", ejaId)
			Run("DELETE FROM ejaTranslations WHERE ejaModuleId=?", ejaId)
			Run("DELETE FROM ejaModuleLinks WHERE dstModuleId=?", ejaId)
		}
	}

	// Delete the entry from the module table
	query := fmt.Sprintf("DELETE FROM %s WHERE ejaId=? AND ejaOwner IN (%s)", moduleName, csv)
	if _, err := Run(query, ejaId); err != nil {
		return err
	}

	// Delete related entries from 'ejaLinks' table
	query = fmt.Sprintf("DELETE FROM ejaLinks WHERE (dstModuleId=? AND dstFieldId=?) OR (srcModuleId=? AND srcFieldId=?) AND ejaOwner IN (%s)", csv)
	if _, err := Run(query, moduleId, ejaId, moduleId, ejaId); err != nil {
		return err
	}

	return nil
}
