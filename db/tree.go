// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
)

type TypeModuleTree struct {
	Id    int64
	Name  string
	Label string
}

func (session *TypeSession) ModuleTree(ownerId int64, moduleId int64, modulePath []TypeModulePath) (result []TypeModuleTree) {
	ejaPermissions := session.ModuleGetIdByName("ejaPermissions")
	ejaUsers := session.ModuleGetIdByName("ejaUsers")
	ejaGroups := session.ModuleGetIdByName("ejaGroups")

	owners := session.NumbersToCsv(session.UserGroupList(ownerId))

	rows, err := session.Rows("SELECT ejaId, name FROM ejaModules WHERE parentId=? ORDER BY power ASC", moduleId)
	if err != nil {
		return
	}

	if len(rows) == 0 {
		if len(modulePath) == 0 {
			rows, err = session.Rows("SELECT ejaId, name FROM ejaModules WHERE parentId=0 AND ejaId!=? ORDER BY power ASC", moduleId)
			if err != nil {
				return
			}
		}
	}

	if len(rows) > 0 {
		for _, row := range rows {
			query := fmt.Sprintf(`
				SELECT ejaId
				FROM ejaLinks
				WHERE srcModuleId = ?
    			AND srcFieldId IN (
        		SELECT ejaId
        		FROM ejaPermissions
        		WHERE ejaModuleId = ?
    			)
    			AND (
        		(dstFieldId = ? AND dstModuleId = ?)
        		OR (dstModuleId = ? AND dstFieldId IN (%s))
    			)
				LIMIT 1
			`, owners)
			checkId, _ := session.Value(query, ejaPermissions, session.Number(row["ejaId"]), ownerId, ejaUsers, ejaGroups)
			if ownerId == 1 || session.Number(checkId) > 0 {
				if !session.IsSubModule(session.Number(row["ejaId"])) {
					result = append(result, TypeModuleTree{Id: session.Number(row["ejaId"]), Name: row["name"], Label: session.Translate(row["name"], ownerId)})
				}
			}
		}
	}

	return
}
