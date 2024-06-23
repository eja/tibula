// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eja/tibula/log"
)

// ModuleAppend appends data to an existing module
func ModuleAppend(module TypeModule, moduleName string) error {
	const owner = 1
	if moduleName == "" {
		moduleName = String(module.Name)
	}
	moduleId := ModuleGetIdByName(moduleName)

	if moduleId < 1 {
		err := fmt.Errorf("Cannot append data, module does not exists")
		log.Error("[db]", err)
		return err
	} else {
		for _, data := range module.Data {
			var keys = []string{"ejaLog"}
			var values = []string{"?"}
			var args = []interface{}{Now()}
			for key, val := range data {
				keys = append(keys, key)
				values = append(values, "?")
				args = append(args, val)
			}
			query := fmt.Sprintf("INSERT INTO %s (ejaId, ejaOwner, %s) VALUES (NULL,1,%s)", moduleName, strings.Join(keys, ", "), strings.Join(values, ","))
			if _, err := Run(query, args...); err != nil {
				log.Error("[db]", "data append", err)
			}
		}
	}
	return nil
}

// ModuleImport imports a module into the database based on the provided TypeModule and module name.
func ModuleImport(module TypeModule, moduleName string) error {
	const owner = 1
	if moduleName == "" {
		moduleName = String(module.Name)
	}

	moduleId := ModuleGetIdByName(moduleName)

	if moduleId < 1 {
		moduleIdRun, err := Run(`
			INSERT INTO ejaModules 
				(ejaId, ejaOwner, ejaLog, name, power, searchLimit, sqlCreated, sortList, parentId) 
      VALUES 
				(NULL,?,?,?,?,?,?,?,?)
			`, owner, Now(), moduleName,
			module.Module.Power,
			module.Module.SearchLimit,
			module.Module.SqlCreated,
			module.Module.SortList,
			ModuleGetIdByName(module.Module.ParentName),
		)
		if err != nil {
			return err
		}
		moduleId = moduleIdRun.LastId
		if err := TableAdd(moduleName); err != nil {
			return err
		}

	}

	if moduleId > 0 {
		_, err := Run(`DELETE FROM ejaFields WHERE ejaModuleId=?`, moduleId)
		if err != nil {
			return err
		}

		for _, field := range module.Field {
			if module.Module.SqlCreated > 0 {
				if check, err := FieldExists(moduleName, field.Name); !check {
					if err != nil {
						return err
					}
					if err := FieldAdd(moduleName, field.Name, field.Type); err != nil {
						return err
					}
				}
			}
			run, err := Run(`
					INSERT INTO ejaFields 
						(ejaId, ejaOwner, ejaLog, ejaModuleId, name, type, value, translate, powerSearch, powerList, powerEdit) 
          VALUES 
						(NULL,?,?,?,?,?,?,?,?,?,?)
					`, owner, Now(), moduleId, field.Name, field.Type, field.Value, field.Translate, field.PowerSearch, field.PowerList, field.PowerEdit)
			if err != nil {
				return err
			}
			Run(`UPDATE ejaFields SET sizeSearch=? WHERE ejaId=?`, field.SizeSearch, run.LastId)
			Run(`UPDATE ejaFields SET sizeList=? WHERE ejaId=?`, field.SizeList, run.LastId)
			Run(`UPDATE ejaFields SET sizeEdit=? WHERE ejaId=?`, field.SizeEdit, run.LastId)
		}

		ejaPermissionsId := ModuleGetIdByName("ejaPermissions")
		ejaUsersId := ModuleGetIdByName("ejaUsers")

		_, err = Run(`
			DELETE FROM ejaLinks 
			WHERE dstModuleId=? 
				AND srcModuleId=? 
				AND srcFieldId IN (SELECT t.ejaId FROM ejaPermissions AS t WHERE t.ejaModuleId=?)
			`, ejaUsersId, ejaPermissionsId, moduleId,
		)
		if err != nil {
			return err
		}

		_, err = Run(`DELETE FROM ejaPermissions WHERE ejaModuleId=?`, moduleId)
		if err != nil {
			return err
		}

		for _, command := range module.Command {
			run, err := Run(`
				INSERT INTO ejaPermissions 
					(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaCommandId) 
				VALUES 
					(NULL,?,?,?,(SELECT t.ejaId FROM ejaCommands AS t WHERE t.name=? LIMIT 1))
				`, owner, Now(), moduleId, command)
			if err != nil {
				return err
			}
			id := run.LastId

			if id > 0 {
				_, err := Run(`
					INSERT INTO ejaLinks 
						(ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power) 
          VALUES 
						(NULL,?,?,?,?,?,?,1)
					`, owner, Now(), ejaPermissionsId, id, ejaUsersId, owner)
				if err != nil {
					return err
				}
			}
		}

		_, err = Run(`DELETE FROM ejaTranslations WHERE ejaModuleId=?`, moduleId)
		if err != nil {
			return err
		}

		_, err = Run(`DELETE FROM ejaTranslations WHERE word=? AND ejaModuleId < 1`, moduleName)
		if err != nil {
			return err
		}

		for _, field := range module.Translation {
			moduleTmpId := moduleId
			if field.EjaModuleName != moduleName {
				moduleTmpId = 0
			}

			_, err := Run(`
				INSERT INTO ejaTranslations 
					(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaLanguage, word, translation) 
        VALUES 
					(NULL,?,?,?,?,?,?)
				`, owner, Now(), moduleTmpId, field.EjaLanguage, field.Word, field.Translation)
			if err != nil {
				return err
			}
		}

		for _, data := range module.Data {
			var keys = []string{"ejaLog"}
			var values = []string{"?"}
			var args = []interface{}{Now()}
			for key, val := range data {
				keys = append(keys, key)
				values = append(values, "?")
				args = append(args, val)
			}
			query := fmt.Sprintf("INSERT INTO %s (ejaId, ejaOwner, %s) VALUES (NULL,1,%s)", moduleName, strings.Join(keys, ", "), strings.Join(values, ","))
			if _, err := Run(query, args...); err != nil {
				log.Error("[db]", "module import", err)
			}
		}

		return nil
	}

	return errors.New("cannot import module")
}
