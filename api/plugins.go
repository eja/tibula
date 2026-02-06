// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
)

type TypePlugins map[string]func(Api, DbSession) Api

var Plugins = TypePlugins{
	"ejaProfile": func(eja Api, db DbSession) Api {
		eja.Alert = nil
		if eja.Action == "run" {
			v := eja.Values
			if v["passwordOld"] == "" || v["passwordNew"] == "" || v["passwordRepeat"] == "" {
				eja.alert(db.Translate("passwordEmptyError", eja.Owner))
			} else if v["passwordNew"] != v["passwordRepeat"] {
				eja.alert(db.Translate("passwordMatchError", eja.Owner))
			} else {
				user := db.UserGetAllBySession(eja.Session)
				if db.Password(v["passwordOld"]) != user["password"] {
					eja.alert(db.Translate("passwordOldError", eja.Owner))
				} else {
					_, err := db.Run("UPDATE ejaUsers SET password=? WHERE ejaId=?", db.Password(v["passwordNew"]), eja.Owner)
					if err == nil {
						eja.info(db.Translate("passwordUpdated", eja.Owner))
					}
				}
			}
		}
		return eja
	},
	"ejaModuleImport": func(eja Api, db DbSession) Api {
		if eja.Action == "run" {
			var module DbModule
			if err := json.Unmarshal([]byte(eja.Values["import"]), &module); err != nil {
				eja.alert(db.Translate("ejaImportJsonError", eja.Owner))
			} else {
				var err error
				dImp := db.Number(eja.Values["dataImport"])
				if dImp == 2 {
					err = db.ModuleAppend(module, eja.Values["moduleName"])
				} else {
					if dImp < 1 {
						module.Data = nil
					}
					err = db.ModuleImport(module, eja.Values["moduleName"])
				}
				if err != nil {
					eja.alert(db.Translate("ejaImportError", eja.Owner))
				} else {
					eja.Values["import"] = ""
					eja.info(db.Translate("ejaImportOk", eja.Owner))
				}
			}
		}
		return eja
	},
	"ejaModuleExport": func(eja Api, db DbSession) Api {
		if eja.Action == "run" {
			mId := db.Number(eja.Values["ejaModuleId"])
			dExp := db.Number(eja.Values["dataExport"]) > 0
			if data, err := db.ModuleExport(mId, dExp); err != nil {
				eja.alert(db.Translate("ejaExportError", eja.Owner))
			} else {
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				eja.Values["export"] = string(jsonData)
				eja.info(db.Translate("ejaExportOk", eja.Owner))
			}
		}
		return eja
	},
	"ejaGroupImport": func(eja Api, db DbSession) Api {
		if eja.Action == "run" {
			var group DbGroup
			if err := json.Unmarshal([]byte(eja.Values["import"]), &group); err != nil {
				eja.alert(db.Translate("ejaImportJsonError", eja.Owner))
			} else if err := db.GroupImport(group, eja.Values["groupName"]); err != nil {
				eja.alert(db.Translate("ejaImportError", eja.Owner))
			} else {
				eja.Values["import"] = ""
				eja.info(db.Translate("ejaImportOk", eja.Owner))
			}
		}
		return eja
	},
	"ejaGroupExport": func(eja Api, db DbSession) Api {
		if eja.Action == "run" {
			gId := db.Number(eja.Values["ejaGroupId"])
			if data, err := db.GroupExport(gId); err != nil {
				eja.alert(db.Translate("ejaExportError", eja.Owner))
			} else {
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				eja.Values["export"] = string(jsonData)
				eja.info(db.Translate("ejaExportOk", eja.Owner))
			}
		}
		return eja
	},
}
