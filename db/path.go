// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
)

// TypeModulePath represents a node in the module hierarchy with its ID, name, and translated label.
type TypeModulePath struct {
	Id    int64
	Name  string
	Label string
}

// ModulePath retrieves the path of module hierarchy from the specified module to its root for a given user.
func ModulePath(ownerId int64, moduleId int64) (result []TypeModulePath) {
	id := moduleId
	ejaPermissions := ModuleGetIdByName("ejaPermissions")
	ejaUsers := ModuleGetIdByName("ejaUsers")
	ejaGroups := ModuleGetIdByName("ejaGroups")
	owners := NumbersToCsv(UserGroupList(ownerId))

	for id != 0 {
		row, _ := Row("SELECT ejaId, parentId, name FROM ejaModules WHERE ejaId=?", id)
		result = append(result, TypeModulePath{
			Id:    Number(row["ejaId"]),
			Name:  row["name"],
			Label: Translate(row["name"], ownerId),
		})
		id = 0
		if len(row) > 0 {
			query := fmt.Sprintf(`
				SELECT ejaId FROM ejaLinks 
				WHERE 
					srcModuleId=? 
					AND srcFieldId IN (SELECT ejaId FROM ejaPermissions WHERE ejaModuleId=?) 
					AND ((dstFieldId=? AND dstModuleId=?) || (dstModuleId=? AND dstFieldId IN (%s)))
				LIMIT 1
				`, owners)
			checkId, _ := Value(query, ejaPermissions, Number(row["ejaId"]), ownerId, ejaUsers, ejaGroups)
			if (ownerId == 1 || Number(checkId) > 0) && Number(row["parentId"]) > 0 {
				id = Number(row["parentId"])
			}
		}
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	result = result[:len(result)-1]

	return
}
