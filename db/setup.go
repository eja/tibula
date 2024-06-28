// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:assets
var Assets embed.FS

// Setup initializes the database with modules, fields, and commands.
// It reads JSON files from the specified setupPath or embeded assets, and populates the database accordingly.
// The admin user credentials are used for setup.
func (session *TypeSession) Setup(setupPath string) error {
	moduleIdMap := map[string]int64{}
	var modules []TypeModule
	var files []string
	var err error

	if setupPath != "" {
		err := filepath.WalkDir(setupPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		entries, err := fs.ReadDir(Assets, "assets")
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, "assets/"+entry.Name())
			}
		}
	}

	for _, file := range files {
		if filepath.Ext(file) == ".json" {
			var fileContent []byte
			if setupPath != "" {
				fileContent, err = os.ReadFile(file)
			} else {
				fileContent, err = Assets.ReadFile(file)
			}
			if err != nil {
				return err
			}

			var module TypeModule
			if err := json.Unmarshal(fileContent, &module); err != nil {
				return err
			}

			if err := session.TableAdd(module.Name); err != nil {
				return err
			}

			for _, field := range module.Field {
				switch field.Name {
				case "ejaId", "ejaOwner", "ejaLog":
					continue
				default:
					if err := session.FieldAdd(module.Name, field.Name, field.Type); err != nil {
						return err
					}
				}
			}

			// add commands
			if module.Name == "ejaCommands" {
				for _, data := range module.Data {
					_, err := session.Run(
						"INSERT INTO ejaCommands (ejaId, ejaOwner, ejaLog, name, powerSearch, powerList, powerEdit, linking, defaultCommand) VALUES (NULL,1,?,?,?,?,?,?,?)",
						session.Now(),
						data["name"],
						data["powerSearch"],
						data["powerList"],
						data["powerEdit"],
						data["linking"],
						data["defaultCommand"],
					)
					if err != nil {
						return err
					}
				}
				module.Data = nil
			}

			modules = append(modules, module)
		}
	}

	// add modules
	for _, module := range modules {
		_, err = session.Run(
			"INSERT INTO ejaModules (ejaId, ejaOwner, ejaLog, name, power, searchLimit, sqlCreated, sortList, parentId) VALUES (NULL, 1, ?, ?, ?, ?, ?, ?, 0)",
			session.Now(),
			module.Name,
			module.Module.Power,
			module.Module.SearchLimit,
			module.Module.SqlCreated,
			module.Module.SortList,
		)
		if err != nil {
			return err
		}
	}
	for _, module := range modules {
		moduleParentId := session.ModuleGetIdByName(module.Module.ParentName)
		moduleId := session.ModuleGetIdByName(module.Name)
		if moduleId > 0 {
			moduleIdMap[module.Name] = moduleId
			if moduleParentId > 0 {
				_, err := session.Run("UPDATE ejaModules SET parentId=? WHERE ejaId=?", moduleParentId, moduleId)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, module := range modules {
		if err := session.ModuleImport(module, module.Name); err != nil {
			return err
		}
	}

	// add module links
	if moduleIdMap["ejaGroups"] > 0 {
		if moduleIdMap["ejaPermissions"] > 0 {
			if _, err := session.Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", session.Now(), moduleIdMap["ejaGroups"], moduleIdMap["ejaPermissions"], 2); err != nil {
				return err
			}
		}
		if moduleIdMap["ejaModules"] > 0 {
			if _, err := session.Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", session.Now(), moduleIdMap["ejaGroups"], moduleIdMap["ejaModules"], 1); err != nil {
				return err
			}
		}
	}
	if moduleIdMap["ejaUsers"] > 0 {
		if moduleIdMap["ejaGroups"] > 0 {
			if _, err := session.Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", session.Now(), moduleIdMap["ejaUsers"], moduleIdMap["ejaGroups"], 1); err != nil {
				return err
			}
		}
		if moduleIdMap["ejaPermissions"] > 0 {
			if _, err := session.Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", session.Now(), moduleIdMap["ejaUsers"], moduleIdMap["ejaPermissions"], 2); err != nil {
				return err
			}
		}
	}
	return nil
}

func (session *TypeSession) SetupAdmin(setupUser string, setupPass string) error {
	if setupPass == "" {
		return errors.New("password is mandatory")
	} else {
		session.Run("DELETE FROM ejaUsers WHERE ejaId=1")
		if _, err := session.Run("INSERT INTO ejaUsers (ejaOwner,ejaLog,username,password,defaultModuleId,ejaLanguage) VALUES (1,?,?,?,?,?)",
			session.Now(),
			setupUser,
			session.Password(setupPass),
			session.ModuleGetIdByName("eja"),
			"en",
		); err != nil {
			return err
		}
	}

	return nil
}
