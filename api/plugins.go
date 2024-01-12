// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
	"github.com/eja/tibula/db"
)

type TypePlugins map[string]func(TypeApi) TypeApi

var Plugins = TypePlugins{
	"ejaProfile": func(eja TypeApi) TypeApi {
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
					db.Put(eja.Owner, db.ModuleGetIdByName("ejaUsers"), db.Number(user["ejaId"]), "password", db.Password(eja.Values["passwordNew"]))
					info(&eja.Info, db.Translate("passwordUpdated", eja.Owner))
				}
			}
		}
		return eja
	},
	"ejaImport": func(eja TypeApi) TypeApi {
		if eja.Action == "run" {
			moduleName := eja.Values["moduleName"]
			moduleData := eja.Values["import"]
			dataImport := db.Number(eja.Values["dataImport"]) > 0
			if moduleData != "" {
				var module db.TypeModule
				if err := json.Unmarshal([]byte(moduleData), &module); err != nil {
					alert(&eja.Alert, db.Translate("ejaImportJsonError", eja.Owner))
				} else {
					if !dataImport {
						module.Data = nil
					}
					if err := db.ModuleImport(module, moduleName); err != nil {
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
	"ejaExport": func(eja TypeApi) TypeApi {
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
}
