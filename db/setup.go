// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/term"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:assets
var assets embed.FS

// Setup initializes the database with modules, fields, and commands.
// It reads JSON files from the specified setupPath or embeded assets, and populates the database accordingly.
// The admin user credentials are used for setup.
func Setup(setupPath string, setupUser string, setupPass string) error {
	var modules []TypeModule
	var files []string
	var err error

	if setupPass == "" {
		fd := int(os.Stdin.Fd())
		if term.IsTerminal(fd) {
			if setupUser == "admin" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Username (admin): ")
				user, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				user = strings.TrimSpace(user)
				if user != "" {
					setupUser = user
				}
			}
			fmt.Print("Password: ")
			if pass, err := term.ReadPassword(fd); err != nil {
				return err
			} else if len(pass) == 0 {
				fmt.Println()
				return errors.New("Password cannot be empty")
			} else {
				fmt.Printf("\nRepeat password: ")
				if passCheck, err := term.ReadPassword(fd); err != nil {
					return err
				} else {
					fmt.Println()
					if !bytes.Equal(pass, passCheck) {
						return errors.New("Passwords do not match")
					}
				}
				setupPass = string(pass)
			}
		}
	}
	if setupPass == "" {
		return errors.New("Setup admin user/pass are mandatory")
	}

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
		entries, err := fs.ReadDir(assets, "assets")
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
				fileContent, err = assets.ReadFile(file)
			}
			if err != nil {
				return err
			}

			var module TypeModule
			if err := json.Unmarshal(fileContent, &module); err != nil {
				return err
			}

			if err := TableAdd(module.Name); err != nil {
				return err
			}

			for _, field := range module.Field {
				switch field.Name {
				case "ejaId", "ejaOwner", "ejaLog":
					continue
				default:
					if err := FieldAdd(module.Name, field.Name, field.Type); err != nil {
						return err
					}
				}
			}

			// add commands
			if module.Name == "ejaCommands" {
				for _, data := range module.Data {
					_, err := Run(
						"INSERT INTO ejaCommands (ejaId, ejaOwner, ejaLog, name, powerSearch, powerList, powerEdit, linking, defaultCommand) VALUES (NULL,1,?,?,?,?,?,?,?)",
						Now(),
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
		_, err = Run(
			"INSERT INTO ejaModules (ejaId, ejaOwner, ejaLog, name, power, searchLimit, sqlCreated, sortList, parentId) VALUES (NULL, 1, ?, ?, ?, ?, ?, ?, 0)",
			Now(),
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
		moduleParentId := ModuleGetIdByName(module.Module.ParentName)
		moduleId := ModuleGetIdByName(module.Name)
		if moduleId > 0 && moduleParentId > 0 {
			_, err := Run("UPDATE ejaModules SET parentId=? WHERE ejaId=?", moduleParentId, moduleId)
			if err != nil {
				return err
			}
		}
	}
	for _, module := range modules {
		if err := ModuleImport(module, module.Name); err != nil {
			return err
		}
	}

	// add module links
	if _, err := Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", Now(), ModuleGetIdByName("ejaGroups"), ModuleGetIdByName("ejaPermissions"), 2); err != nil {
		return err
	}
	if _, err := Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", Now(), ModuleGetIdByName("ejaGroups"), ModuleGetIdByName("ejaModules"), 1); err != nil {
		return err
	}
	if _, err := Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", Now(), ModuleGetIdByName("ejaUsers"), ModuleGetIdByName("ejaGroups"), 1); err != nil {
		return err
	}
	if _, err := Run("INSERT INTO ejaModuleLinks (ejaOwner,ejaLog,dstModuleId,srcModuleId,power) VALUES (1,?,?,?,?)", Now(), ModuleGetIdByName("ejaUsers"), ModuleGetIdByName("ejaPermissions"), 2); err != nil {
		return err
	}

	// add admin user
	if _, err := Run("INSERT INTO ejaUsers (ejaOwner,ejaLog,username,password,defaultModuleId,ejaLanguage) VALUES (1,?,?,?,?,?)", Now(), setupUser, Password(setupPass), ModuleGetIdByName("eja"), "en"); err != nil {
		return err
	}

	return nil
}
