// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
)

type TypePlugins map[string]func(TypeApi, TypeDbSession) TypeApi

var Plugins = TypePlugins{
	"ejaProfile": func(eja TypeApi, db TypeDbSession) TypeApi {
		eja.Alert = nil
		if eja.Action == "run" {
			if eja.Values["passwordOld"] == "" || eja.Values["passwordNew"] == "" || eja.Values["passwordRepeat"] == "" {
				alert(&eja.Alert, db.Translate("passwordEmptyError", eja.Owner))
			} else if eja.Values["passwordNew"] != eja.Values["passwordRepeat"] {
				alert(&eja.Alert, db.Translate("passwordMatchError", eja.Owner))
			} else {
				user := db.UserGetAllBySession(eja.Session)
				pass := db.Password(eja.Values["passwordOld"])
				if pass != user["password"] {
					alert(&eja.Alert, db.Translate("passwordOldError", eja.Owner))
				} else {
					_, err := db.Run("UPDATE ejaUsers SET password=? WHERE ejaId=?", db.Password(eja.Values["passwordNew"]), eja.Owner)
					if err == nil {
						info(&eja.Info, db.Translate("passwordUpdated", eja.Owner))
					}
				}
			}
		}
		return eja
	},
	"ejaModuleImport": func(eja TypeApi, db TypeDbSession) TypeApi {
		if eja.Action == "run" {
			moduleName := eja.Values["moduleName"]
			moduleData := eja.Values["import"]
			dataImport := db.Number(eja.Values["dataImport"])
			if moduleData != "" {
				var module TypeDbModule
				if err := json.Unmarshal([]byte(moduleData), &module); err != nil {
					alert(&eja.Alert, db.Translate("ejaImportJsonError", eja.Owner))
				} else {
					var err error
					if dataImport == 2 {
						err = db.ModuleAppend(module, moduleName)
					} else {
						if dataImport < 1 {
							module.Data = nil
						}
						err = db.ModuleImport(module, moduleName)
					}
					if err != nil {
						alert(&eja.Alert, db.Translate("ejaImportError", eja.Owner))
					} else {
						eja.Values["import"] = ""
						info(&eja.Info, db.Translate("ejaImportOk", eja.Owner))
					}
				}
			}
		}
		return eja
	},
	"ejaModuleExport": func(eja TypeApi, db TypeDbSession) TypeApi {
		if eja.Action == "run" {
			moduleId := db.Number(eja.Values["ejaModuleId"])
			dataExport := db.Number(eja.Values["dataExport"]) > 0
			if moduleId > 0 {
				if data, err := db.ModuleExport(moduleId, dataExport); err != nil {
					alert(&eja.Alert, db.Translate("ejaExportError", eja.Owner))
				} else {
					jsonData, _ := json.MarshalIndent(data, "", "  ")
					eja.Values["export"] = string(jsonData)
					info(&eja.Info, db.Translate("ejaExportOk", eja.Owner))
				}
			}
		}
		return eja
	},
	"ejaGroupImport": func(eja TypeApi, db TypeDbSession) TypeApi {
		if eja.Action == "run" {
			groupName := eja.Values["groupName"]
			groupData := eja.Values["import"]
			if groupData != "" {
				var group TypeDbGroup
				if err := json.Unmarshal([]byte(groupData), &group); err != nil {
					alert(&eja.Alert, db.Translate("ejaImportJsonError", eja.Owner))
				} else {
					var err error
					err = db.GroupImport(group, groupName)
					if err != nil {
						alert(&eja.Alert, db.Translate("ejaImportError", eja.Owner))
					} else {
						eja.Values["import"] = ""
						info(&eja.Info, db.Translate("ejaImportOk", eja.Owner))
					}
				}
			}
		}
		return eja
	},
	"ejaGroupExport": func(eja TypeApi, db TypeDbSession) TypeApi {
		if eja.Action == "run" {
			groupId := db.Number(eja.Values["ejaGroupId"])
			if groupId > 0 {
				if data, err := db.GroupExport(groupId); err != nil {
					alert(&eja.Alert, db.Translate("ejaExportError", eja.Owner))
				} else {
					jsonData, _ := json.MarshalIndent(data, "", "  ")
					eja.Values["export"] = string(jsonData)
					info(&eja.Info, db.Translate("ejaExportOk", eja.Owner))
				}
			}
		}
		return eja
	},
}
