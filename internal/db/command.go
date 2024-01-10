// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

// TypeCommand represents a command with its properties.
type TypeCommand struct {
	Name   string
	Label  string
	Linker bool
}

// Commands retrieves a list of commands based on user and module information.
// It checks permissions and builds the command list accordingly.
func Commands(userId int64, moduleId int64, actionType string) ([]TypeCommand, error) {
	commandList := []TypeCommand{}
	actionTypeSql := ""

	moduleName := ModuleGetNameById(moduleId)
	if moduleName == "" {
		return nil, errors.New("module does not exist")
	}

	if moduleName == "ejaLogin" {
		commandList = append(commandList, TypeCommand{Name: "login", Label: Translate("login", userId)})
	}

	if actionType != "" {
		actionTypeSql = fmt.Sprintf(" AND power%s > 0 ORDER BY power%s ASC ", actionType, actionType)
	}

	query := fmt.Sprintf(`
    SELECT *
    FROM ejaCommands
    WHERE ejaId IN (
        SELECT ejaCommandId
        FROM ejaPermissions
        WHERE ejaModuleId=? AND ejaId IN (
            SELECT srcFieldId
            FROM ejaLinks
            WHERE srcModuleId=? AND (
                (dstModuleId=? AND dstFieldId=?)
                OR (dstModuleId=? AND dstFieldId IN (%s))
            )
        )
    ) %s`,
		UserGroupCsv(userId), actionTypeSql)

	rows, err := Rows(
		query,
		moduleId,
		ModuleGetIdByName("ejaPermissions"),
		ModuleGetIdByName("ejaUsers"),
		userId,
		ModuleGetIdByName("ejaGroups"),
	)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		commandList = append(commandList, TypeCommand{Name: row["name"], Label: Translate(row["name"], userId), Linker: Number(row["linking"]) > 0})
	}
	return commandList, nil
}

// CommandExists checks if a given command exists in the provided list of commands.
func CommandExists(commands []TypeCommand, commandName string) bool {
	for _, row := range commands {
		if row.Name == commandName {
			return true
		}
	}
	return false
}
