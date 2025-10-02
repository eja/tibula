// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

type TypeCommand struct {
	Name   string
	Label  string
	Linker bool
}

func (session *TypeSession) Commands(userId int64, moduleId int64, actionType string) ([]TypeCommand, error) {
	commandList := []TypeCommand{}
	actionTypeSql := ""

	moduleName := session.ModuleGetNameById(moduleId)
	if moduleName == "" {
		return nil, errors.New("module does not exist")
	}

	if moduleName == "ejaLogin" {
		commandList = append(commandList, TypeCommand{Name: "login", Label: session.Translate("login", userId)})
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
		session.UserGroupCsv(userId), actionTypeSql)

	rows, err := session.Rows(
		query,
		moduleId,
		session.ModuleGetIdByName("ejaPermissions"),
		session.ModuleGetIdByName("ejaUsers"),
		userId,
		session.ModuleGetIdByName("ejaGroups"),
	)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		commandList = append(commandList, TypeCommand{Name: row["name"], Label: session.Translate(row["name"], userId), Linker: session.Number(row["linking"]) > 0})
	}
	return commandList, nil
}

func (session *TypeSession) CommandExists(commands []TypeCommand, commandName string) bool {
	for _, row := range commands {
		if row.Name == commandName {
			return true
		}
	}
	return false
}
