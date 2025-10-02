// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"

	"github.com/eja/tibula/log"
)

func (session *TypeSession) GroupImport(group TypeGroup, groupName string) (err error) {
	const owner = 1

	if group.Type != "group" {
		return errors.New("Wrong module type")
	}

	if groupName == "" {
		groupName = group.Name
		if groupName == "" {
			return errors.New("invalid group name")
		}
	}

	var groupId int64
	groupModuleId := session.ModuleGetIdByName("ejaGroups")
	shareModuleId := session.ModuleGetIdByName("ejaModules")
	permissionModuleId := session.ModuleGetIdByName("ejaPermissions")
	groupId, err = session.New(owner, groupModuleId)
	if err != nil {
		return
	}
	if err = session.Put(owner, groupModuleId, groupId, "name", groupName); err != nil {
		return
	}

	for _, share := range group.Shares {
		_, err := session.Run(`
    	INSERT INTO ejaLinks 
    		(ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power) 
    	VALUES 
      	(NULL,?,?,?,(SELECT lf.ejaId FROM ejaModules AS lf WHERE lf.name=? LIMIT 1),?,?,1)
    	`, owner, session.Now(), shareModuleId, share, groupModuleId, groupId)
		if err != nil {
			return err
		}
	}
	for moduleName, commands := range group.Permissions {
		for _, commandName := range commands {
			_, err := session.Run(`
				INSERT INTO ejaLinks
			  	(ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power)
			  VALUES
					(NULL,?,?,?,
						(
							SELECT lf.ejaId FROM ejaPermissions AS lf 
							WHERE 
								lf.ejaModuleId=(SELECT m.ejaId FROM ejaModules AS m WHERE m.name=? LIMIT 1) 
							AND
								lf.ejaCommandId=(SELECT c.ejaId FROM ejaCommands AS c WHERE c.name=? LIMIT 1)
							LIMIT 1
						) ,?,?,1)
			     	`, owner, session.Now(), permissionModuleId, moduleName, commandName, groupModuleId, groupId)
			if err != nil {
				return err
			}
		}
	}

	return
}

func (session *TypeSession) ModuleAppend(module TypeModule, moduleName string) error {
	const owner = 1

	if module.Type != "module" {
		return errors.New("Wrong module type")
	}

	if moduleName == "" {
		moduleName = session.String(module.Name)
	}
	moduleId := session.ModuleGetIdByName(moduleName)

	if moduleId < 1 {
		err := fmt.Errorf("Cannot append data, module does not exists")
		log.Error(tag, err)
		return err
	} else {
		for _, data := range module.Data {
			if id, err := session.New(owner, moduleId); err != nil {
				log.Error(tag, "data append", err)
			} else {
				for key, val := range data {
					session.Put(owner, moduleId, session.Number(id), key, session.String(val))
				}
			}
		}
	}

	return nil
}

func (session *TypeSession) ModuleImport(module TypeModule, moduleName string) error {
	const owner = 1

	if module.Type != "module" {
		return errors.New("Wrong module type")
	}

	if moduleName == "" {
		moduleName = session.String(module.Name)
	}

	moduleId := session.ModuleGetIdByName(moduleName)

	if moduleId < 1 {
		moduleIdRun, err := session.Run(`
			INSERT INTO ejaModules 
				(ejaId, ejaOwner, ejaLog, name, power, searchLimit, sqlCreated, sortList, parentId) 
      VALUES 
				(NULL,?,?,?,?,?,?,?,?)
			`, owner, session.Now(), moduleName,
			module.Module.Power,
			module.Module.SearchLimit,
			module.Module.SqlCreated,
			module.Module.SortList,
			session.ModuleGetIdByName(module.Module.ParentName),
		)
		if err != nil {
			return err
		}
		moduleId = moduleIdRun.LastId
		if err := session.TableAdd(moduleName); err != nil {
			return err
		}

	}

	if moduleId > 0 {
		_, err := session.Run(`DELETE FROM ejaFields WHERE ejaModuleId=?`, moduleId)
		if err != nil {
			return err
		}

		for _, field := range module.Field {
			if module.Module.SqlCreated > 0 {
				if check, err := session.FieldExists(moduleName, field.Name); !check {
					if err != nil {
						return err
					}
					if err := session.FieldAdd(moduleName, field.Name, field.Type); err != nil {
						return err
					}
				}
			}
			run, err := session.Run(`
					INSERT INTO ejaFields 
						(ejaId, ejaOwner, ejaLog, ejaModuleId, name, type, value, translate, powerSearch, powerList, powerEdit) 
          VALUES 
						(NULL,?,?,?,?,?,?,?,?,?,?)
					`, owner, session.Now(), moduleId, field.Name, field.Type, field.Value, field.Translate, field.PowerSearch, field.PowerList, field.PowerEdit)
			if err != nil {
				return err
			}
			session.Run(`UPDATE ejaFields SET sizeSearch=? WHERE ejaId=?`, field.SizeSearch, run.LastId)
			session.Run(`UPDATE ejaFields SET sizeList=? WHERE ejaId=?`, field.SizeList, run.LastId)
			session.Run(`UPDATE ejaFields SET sizeEdit=? WHERE ejaId=?`, field.SizeEdit, run.LastId)
		}

		ejaPermissionsId := session.ModuleGetIdByName("ejaPermissions")
		ejaUsersId := session.ModuleGetIdByName("ejaUsers")

		_, err = session.Run(`
			DELETE FROM ejaLinks 
			WHERE dstModuleId=? 
				AND srcModuleId=? 
				AND srcFieldId IN (SELECT t.ejaId FROM ejaPermissions AS t WHERE t.ejaModuleId=?)
			`, ejaUsersId, ejaPermissionsId, moduleId,
		)
		if err != nil {
			return err
		}

		_, err = session.Run(`DELETE FROM ejaPermissions WHERE ejaModuleId=?`, moduleId)
		if err != nil {
			return err
		}

		for _, command := range module.Command {
			run, err := session.Run(`
				INSERT INTO ejaPermissions 
					(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaCommandId) 
				VALUES 
					(NULL,?,?,?,(SELECT t.ejaId FROM ejaCommands AS t WHERE t.name=? LIMIT 1))
				`, owner, session.Now(), moduleId, command)
			if err != nil {
				return err
			}
			id := run.LastId

			if id > 0 {
				_, err := session.Run(`
					INSERT INTO ejaLinks 
						(ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power) 
          VALUES 
						(NULL,?,?,?,?,?,?,1)
					`, owner, session.Now(), ejaPermissionsId, id, ejaUsersId, owner)
				if err != nil {
					return err
				}
			}
		}

		_, err = session.Run(`DELETE FROM ejaTranslations WHERE ejaModuleId=?`, moduleId)
		if err != nil {
			return err
		}

		_, err = session.Run(`DELETE FROM ejaTranslations WHERE word=? AND ejaModuleId < 1`, moduleName)
		if err != nil {
			return err
		}

		for _, field := range module.Translation {
			moduleTmpId := moduleId
			if field.EjaModuleName != moduleName {
				moduleTmpId = 0
			}
			_, err := session.Run(`
				INSERT INTO ejaTranslations 
					(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaLanguage, word, translation) 
        VALUES 
					(NULL,?,?,?,?,?,?)
				`, owner, session.Now(), moduleTmpId, field.EjaLanguage, field.Word, field.Translation)
			if err != nil {
				return err
			}
		}

		for _, field := range module.Link {
			srcModuleId := session.ModuleGetIdByName(field.SrcModule)
			dstModuleId := session.ModuleGetIdByName(field.DstModule)
			if srcModuleId > 0 && dstModuleId > 0 {
				alreadyExists, err := session.Value(`SELECT COUNT(*) FROM ejaModuleLinks WHERE srcModuleId=? AND dstModuleId=?`, srcModuleId, dstModuleId)
				if err != nil {
					return err
				}
				if session.Number(alreadyExists) == 0 {
					if _, err := session.Run(`
						INSERT INTO ejaModuleLinks
							(ejaOwner, ejaLog, srcModuleId, srcFieldName, dstModuleId, power)
						VALUES
							(?,?,?,?,?,?);
					`, owner, session.Now(), srcModuleId, field.SrcField, dstModuleId, field.Power); err != nil {
						return err
					}
				}
			}
		}

		for _, data := range module.Data {
			if id, err := session.New(owner, moduleId); err != nil {
				log.Error(tag, "data append", err)
			} else {
				for key, val := range data {
					session.Put(owner, moduleId, session.Number(id), key, session.String(val))
				}
			}
		}

		return nil
	}

	return errors.New("cannot import module")
}
