// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import ()

func (session *TypeSession) IsSubModule(moduleId int64) bool {
	value, err := session.Value(`SELECT srcFieldName FROM ejaModuleLinks WHERE srcModuleId=? AND srcFieldName != "" LIMIT 1`, moduleId)
	if err == nil && value != "" {
		return true
	}
	return false
}

func (session *TypeSession) SubModules(ownerId int64, moduleId int64) (result []TypeLink) {
	ejaPermissions := session.ModuleGetIdByName("ejaPermissions")
	ejaUsers := session.ModuleGetIdByName("ejaUsers")
	rows, err := session.Rows(`
		SELECT srcModuleId, (SELECT name FROM ejaModules WHERE ejaId=srcModuleId) AS srcModuleName 
		FROM ejaModuleLinks 
		WHERE dstModuleId=? 
        AND srcFieldName != ""
		ORDER BY power ASC
		`, moduleId)
	if err != nil {
		return
	}
	for _, row := range rows {
		session.Value("SELECT ejaId FROM ejaLinks WHERE srcModuleId=? AND srcFieldId IN (SELECT ejaId FROM ejaPermissions WHERE ejaModuleId=?) AND dstFieldId=? AND dstModuleId=? LIMIT 1",
			ejaPermissions, session.Number(row["srcModuleId"]), ownerId, ejaUsers)
		result = append(result, TypeLink{
			ModuleId: session.Number(row["srcModuleId"]),
			Label:    session.Translate(row["srcModuleName"], ownerId),
		})
	}
	return
}
