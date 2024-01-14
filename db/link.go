// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import ()

// TypeLink represents a link between modules and fields.
type TypeLink struct {
	Label       string `json:"Label,omitempty"`
	ModuleId    int64  `json:"ModuleId,omitempty"`
	ModuleLabel string `json:"ModuleLabel,omitempty"`
	FieldId     int64  `json:"FieldId,omitempty"`
}

// ModuleLinks retrieves a list of links associated with a specified module.
func ModuleLinks(ownerId int64, moduleId int64) (result []TypeLink) {
	ejaPermissions := ModuleGetIdByName("ejaPermissions")
	ejaUsers := ModuleGetIdByName("ejaUsers")
	rows, err := Rows(`
		SELECT srcModuleId, (SELECT name FROM ejaModules WHERE ejaId=srcModuleId) AS srcModuleName 
		FROM ejaModuleLinks 
		WHERE dstModuleId=? 
		ORDER BY power ASC
		`, moduleId)
	if err != nil {
		return
	}
	for _, row := range rows {
		Value("SELECT ejaId FROM ejaLinks WHERE srcModuleId=? AND srcFieldId IN (SELECT ejaId FROM ejaPermissions WHERE ejaModuleId=?) AND dstFieldId=? AND dstModuleId=? LIMIT 1",
			ejaPermissions, Number(row["srcModuleId"]), ownerId, ejaUsers)
		result = append(result, TypeLink{
			ModuleId: Number(row["srcModuleId"]),
			Label:    Translate(row["srcModuleName"], ownerId),
		})
	}
	return
}

// LinkDel deletes a link between modules and fields.
func LinkDel(ownerId int64, moduleId int64, fieldId int64, linkModuleId int64, linkFieldId int64) error {
	_, err := Run("DELETE FROM ejaLinks WHERE ejaOwner=? AND srcModuleId=? AND srcFieldId=? AND dstModuleId=? AND dstFieldId=?", ownerId, moduleId, fieldId, linkModuleId, linkFieldId)
	return err
}

// LinkAdd adds a new link between modules and fields.
func LinkAdd(ownerId int64, moduleId int64, fieldId int64, linkModuleId int64, linkFieldId int64) error {
	_, err := Run("INSERT INTO ejaLinks (ejaOwner,ejaLog,srcModuleId,srcFieldId,dstModuleId,dstFieldId,power) VALUES(?,?,?,?,?,?,?)", ownerId, Now(), moduleId, fieldId, linkModuleId, linkFieldId, 1)
	return err
}

// LinkCopy duplicates a link from the original field to a new field in a different module.
func LinkCopy(userId int64, dstFieldNew int64, dstModule int64, dstFieldOriginal int64) (TypeRun, error) {
	return Run(`
		INSERT INTO ejaLinks (ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power) 
		SELECT NULL,?,?,srcModuleId,srcFieldId,dstModuleId,?,power 
		FROM ejaLinks 
		WHERE dstModuleId=? AND dstFieldId=?
		`, userId, Now(), dstFieldNew, dstModule, dstFieldOriginal)
}

// SearchLinks searches for links associated with a specified module, field, and owner ID.
func SearchLinks(ownerId int64, srcModuleId int64, srcFieldId int64, dstModuleId int64) []string {
	result := []string{"0"}
	rows, err := Rows("SELECT * FROM ejaLinks WHERE ejaOwner=? AND dstModuleId=? AND dstFieldId=? AND srcModuleId=?", ownerId, srcModuleId, srcFieldId, dstModuleId)
	if err == nil {
		for _, row := range rows {
			result = append(result, row["srcFieldId"])
		}
	}
	return result
}
