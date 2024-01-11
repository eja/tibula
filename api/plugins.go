// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"github.com/eja/tibula/db"
)

type TypePlugins map[string]func(TypeApi) TypeApi

var Plugins = TypePlugins{
	"ejaProfile": func(eja TypeApi) TypeApi {
		eja.Alert = nil
		eja.Commands = append(eja.Commands, db.TypeCommand{Name: "update", Label: db.Translate("update", eja.Owner)})
		if eja.Action == "update" {
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
}
