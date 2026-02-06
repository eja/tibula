// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/eja/tibula/sys"
)

const tag = "[api]"

func Set() Api {
	return Api{
		Language:           sys.Options.Language,
		DefaultSearchLimit: 15,
		DefaultSearchOrder: "ejaId DESC",
		Values:             make(map[string]string),
		SearchOrder:        make(map[string]string),
		Link:               DbLink{},
	}
}

func Run(eja Api, sessionSave bool) (Api, error) {
	db := DbProvider()

	if err := db.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort); err != nil {
		return eja, err
	}
	defer db.Close()

	eja = runAuthPipeline(eja, db)

	if eja.Owner == 0 {
		errName := "ejaNotAuthorized"
		if len(eja.Values) > 0 {
			eja.alert(db.Translate(errName))
		}
		return eja, errors.New(errName)
	}

	if eja.ModuleId == 0 {
		eja.ModuleId = db.ModuleGetIdByName("eja")
	}
	if eja.ModuleName == "" {
		eja.ModuleName = db.ModuleGetNameById(eja.ModuleId)
	}

	eja.Commands, _ = db.Commands(eja.Owner, eja.ModuleId, "")
	if eja.Action != "" && eja.Action != "login" && !db.CommandExists(eja.Commands, eja.Action) {
		eja.alert(db.Translate("ejaNotPermitted", eja.Owner))
		return wrapUp(eja, db, sessionSave), nil
	}

	eja = runStatePipeline(eja, db)

	eja = runDataPipeline(eja, db)

	return wrapUp(eja, db, sessionSave), nil
}

func runAuthPipeline(eja Api, db DbSession) Api {
	var user map[string]string

	switch {
	case eja.Action == "login" && eja.Values["username"] != "" && eja.Values["password"] != "":
		user = db.UserGetAllByUserAndPass(eja.Values["username"], eja.Values["password"])
		if len(user) > 0 {
			eja.Session = db.SessionInit(db.Number(user["ejaId"]))
		}

	case eja.Values["googleSsoToken"] != "":
		if email := googleSsoEmail(eja.Values["googleSsoToken"]); email != "" {
			user = db.UserGetAllByUsername(email)
			if len(user) > 0 {
				eja.Session = db.SessionInit(db.Number(user["ejaId"]))
			}
		}
	}

	if eja.Session != "" {
		if len(user) == 0 {
			user = db.UserGetAllBySession(eja.Session)
			eja.Session = db.SessionTokenUpdate(db.Number(user["ejaId"]), user["ejaSession"])
		}
		if len(user) > 0 {
			eja.Owner = db.Number(user["ejaId"])
			if user["ejaLanguage"] != "" {
				eja.Language = user["ejaLanguage"]
			}
			if eja.ModuleId == 0 && eja.ModuleName != "" {
				eja.ModuleId = db.ModuleGetIdByName(eja.ModuleName)
			}
			if eja.ModuleId == 0 {
				eja.ModuleId = db.Number(user["defaultModuleId"])
				eja.ModuleName = db.ModuleGetNameById(eja.ModuleId)
			}
		}
	}

	if eja.Action == "logout" && eja.Owner > 0 {
		db.SessionReset(eja.Owner)
		eja.Session, eja.Owner = "", 0
	}

	return eja
}

func runStatePipeline(eja Api, db DbSession) Api {
	if eja.SearchLinkClean {
		db.SessionCleanSearch(eja.Owner)
	}

	if rows, err := db.SessionLoad(eja.Owner, eja.ModuleId); err == nil {
		for _, row := range rows {
			switch row["name"] {
			case "SearchLimit":
				eja.SearchLimit = db.Number(row["value"])
			case "SearchOrder":
				if eja.SearchOrder[row["sub"]] == "" {
					eja.SearchOrder[row["sub"]] = row["value"]
				}
			case "SearchOffset":
				eja.SearchOffset = db.Number(row["value"])
			case "SqlQuery64":
				eja.SqlQuery64 = row["value"]
			case "SqlQueryArgs":
				eja.SqlQueryArgs = append(eja.SqlQueryArgs, row["value"])
			case "Link":
				switch row["sub"] {
				case "ModuleId":
					eja.Link.ModuleId = db.Number(row["value"])
				case "FieldId":
					eja.Link.FieldId = db.Number(row["value"])
				case "Label":
					eja.Link.Label = row["value"]
				}
			}
		}
	}

	if eja.Action == "" && db.AutoSearch(eja.ModuleId) {
		eja.ActionType = "List"
		eja.SqlQuery64 = ""
	}
	return eja
}

func runDataPipeline(eja Api, db DbSession) Api {
	var subPath ActiveSubModule
	for i, sub := range eja.SubModulePath {
		eja.SubModulePathString += fmt.Sprintf("%d.%d.%d,", sub.LinkingModuleId, sub.ModuleId, sub.FieldId)
		if i < len(eja.SubModulePath)-1 && eja.ModuleId == eja.SubModulePath[i+1].LinkingModuleId {
			eja.ActionType, eja.Action = "", "edit"
			eja.Id = eja.SubModulePath[i+1].FieldId
			break
		}
		if sub.ModuleId == eja.ModuleId {
			subPath.Found = true
			subPath.Item = sub
			subPath.Item.FieldName = db.ModuleLinksFieldName(sub.ModuleId, sub.LinkingModuleId)
			eja.Values[subPath.Item.FieldName] = db.String(subPath.Item.FieldId)
			break
		}
		if sub.LinkingModuleId == eja.ModuleId {
			eja.Action, eja.Id = "edit", sub.FieldId
		}
	}
	if !subPath.Found {
		eja.SubModulePathString = ""
	}

	linkingField := ""
	if eja.Link.ModuleId > 0 && eja.Link.FieldId > 0 && eja.Link.Label != "" {
		eja.Linking = true
		eja.Link.ModuleLabel = db.Translate(db.ModuleGetNameById(eja.Link.ModuleId), eja.Owner)
		linkingField = db.ModuleLinksFieldName(eja.ModuleId, eja.Link.ModuleId)
		if linkingField != "" && eja.Action != "search" {
			eja.Values[linkingField] = db.String(eja.Link.FieldId)
		}
		if eja.ModuleId == eja.Link.ModuleId && eja.Id == eja.Link.FieldId && eja.Id > 0 {
			db.SessionCleanLink(eja.Owner)
			db.SessionCleanSearch(eja.Owner)
			eja.Link, eja.Action, eja.Linking = DbLink{}, "edit", false
		}
		for _, fid := range eja.IdList {
			if eja.Action == "link" {
				db.LinkDel(eja.Owner, eja.ModuleId, fid, eja.Link.ModuleId, eja.Link.FieldId)
				db.LinkAdd(eja.Owner, eja.ModuleId, fid, eja.Link.ModuleId, eja.Link.FieldId)
			}
			if eja.Action == "unlink" {
				db.LinkDel(eja.Owner, eja.ModuleId, fid, eja.Link.ModuleId, eja.Link.FieldId)
			}
		}
		if eja.Action == "link" || eja.Action == "unlink" {
			eja.IdList, eja.ActionType = []int64{}, "List"
		}
	}

	switch eja.Action {
	case "edit":
		if eja.Id > 0 {
			eja.Values = db.TableGetAllById(eja.ModuleName, eja.Id)
		}
	case "new", "copy":
		oldId := eja.Id
		if nid, err := db.New(eja.Owner, eja.ModuleId); err == nil && nid > 0 {
			eja.Id = nid
			db.LinkCopy(eja.Owner, eja.Id, eja.ModuleId, oldId)
		} else {
			eja.alert(db.Translate("ejaActionNewError", eja.Owner))
		}
	case "delete":
		ids := eja.IdList
		if len(ids) == 0 && eja.Id > 0 {
			ids = []int64{eja.Id}
		}
		for _, vid := range ids {
			err := db.Del(eja.Owner, eja.ModuleId, vid)
			if eja.ModuleName == "ejaModules" {
				msg := "ejaSqlModuleDeleteTrue"
				if err != nil {
					msg = "ejaSqlModuleDeleteFalse"
					eja.alert(db.Translate(msg, eja.Owner))
				} else {
					eja.info(db.Translate(msg, eja.Owner))
				}
			}
		}
		eja.ActionType = "List"
	}

	if len(eja.Values) > 0 && (eja.Action == "save" || eja.Action == "copy" || eja.Action == "new") {
		eja = handleSave(eja, db)
	}

	if eja.Action == "list" && eja.SqlQuery64 == "" {
		eja.Values, eja.Action, eja.Id = make(map[string]string), "", 0
	}

	if contains([]string{"search", "previous", "next", "list"}, eja.Action) || eja.ActionType == "List" {
		eja = handleSearch(eja, db, linkingField, subPath)
	}

	return eja
}

func handleSave(eja Api, db DbSession) Api {
	if eja.ModuleName == "ejaModules" {
		if db.Number(eja.Values["sqlCreated"]) > 0 {
			if err := db.TableAdd(eja.Values["name"]); err != nil {
				eja.alert(db.Translate("ejaSqlModuleNotCreated", eja.Owner))
			} else {
				eja.info(db.Translate("ejaSqlModuleCreated", eja.Owner))
			}
		}
		if eja.Action == "save" && db.PermissionCount(eja.Id) == 0 {
			if db.Number(eja.Values["sqlCreated"]) > 0 {
				db.PermissionAddDefault(eja.Owner, eja.Id)
				eja.info(db.Translate("ejaModulePermissionsAddDefault", eja.Owner))
			} else {
				db.PermissionAdd(eja.Owner, eja.Id, "logout")
				eja.info(db.Translate("ejaModulePermissionAdd", eja.Owner))
			}
			db.UserPermissionCopy(eja.Owner, eja.Id)
		}
	}

	if eja.ModuleName == "ejaFields" {
		moduleName := db.ModuleGetNameById(db.Number(eja.Values["ejaModuleId"]))
		if err := db.FieldAdd(moduleName, eja.Values["name"], eja.Values["type"]); err == nil {
			eja.info(db.Translate("ejaSqlFieldCreated", eja.Owner))
		} else {
			eja.alert(db.Translate("ejaSqlFieldNotCreated", eja.Owner))
		}
	}

	if eja.Id < 1 {
		if nid, err := db.New(eja.Owner, eja.ModuleId); err == nil {
			eja.Id = nid
		}
	}

	if eja.Id < 1 {
		eja.alert(db.Translate("ejaErrorEditId", eja.Owner))
		return eja
	}

	for k, v := range eja.Values {
		var val interface{}
		switch db.FieldTypeGet(eja.ModuleId, k) {
		case "password":
			val = v
			if len(v) != 64 {
				val = db.Password(v)
			}
		case "boolean", "integer":
			val = db.Number(v)
		case "decimal":
			val = db.Float(v)
		default:
			val = db.String(v)
		}

		if k == "ejaOwner" && db.Number(v) < 1 {
			db.Put(eja.Owner, eja.ModuleId, eja.Id, k, eja.Owner)
		} else {
			db.Put(eja.Owner, eja.ModuleId, eja.Id, k, val)
		}
	}

	if res, err := db.Get(eja.Owner, eja.ModuleId, eja.Id); err == nil {
		eja.Values = res
	}
	return eja
}

func handleSearch(eja Api, db DbSession, linkField string, sub ActiveSubModule) Api {
	eja.ActionType = "List"
	mDef := db.TableGetAllById("ejaModules", eja.ModuleId)

	limit := eja.SearchLimit
	if limit < 1 {
		limit = db.Number(mDef["searchLimit"])
	}
	if limit < 1 {
		limit = eja.DefaultSearchLimit
	}
	eja.SearchLimit = limit

	if eja.Action == "previous" && eja.SearchOffset >= limit {
		eja.SearchOffset -= limit
	} else if eja.Action == "next" {
		eja.SearchOffset += limit
	}

	var sqlQuery string
	var sqlArgs []interface{}

	if eja.SqlQuery64 != "" {
		if b, err := base64.StdEncoding.DecodeString(eja.SqlQuery64); err == nil {
			sqlQuery = db.String(b)
		}
	} else {
		var err error
		sqlQuery, sqlArgs, err = db.SearchQuery(eja.Owner, eja.ModuleName, eja.Values)
		if err == nil {
			eja.SqlQuery64 = base64.StdEncoding.EncodeToString([]byte(sqlQuery))
			eja.SqlQueryArgs = sqlArgs
			db.SessionPut(eja.Owner, "SqlQuery64", eja.SqlQuery64)
			for _, arg := range sqlArgs {
				db.SessionPut(eja.Owner, "SqlQueryArgs", db.String(arg))
			}
		}
	}

	var sqlOrder string
	for _, key := range db.FieldNameList(eja.ModuleId, "List") {
		v := eja.SearchOrder[key]
		if v == "ASC" || v == "DESC" {
			db.SessionPut(eja.Owner, "SearchOrder", v, key)
			if sqlOrder != "" {
				sqlOrder += ","
			}
			sqlOrder += key + " " + v
		}
	}
	if sqlOrder == "" {
		sqlOrder = eja.DefaultSearchOrder
		if mDef["sortList"] != "" {
			sqlOrder = mDef["sortList"] + " ASC"
		}
	}

	sqlLinks := ""
	if linkField != "" {
		sqlLinks = fmt.Sprintf(" AND %s=? ", linkField)
		eja.SqlQueryArgs = append(eja.SqlQueryArgs, db.String(eja.Link.FieldId))
	} else if eja.SearchLink {
		sqlLinks = db.SearchQueryLinks(eja.Owner, eja.Link.ModuleId, eja.Link.FieldId, eja.ModuleId)
	}

	if sub.Found && sub.Item.FieldId > 0 {
		sqlQuery += fmt.Sprintf(" AND %s=%d ", sub.Item.FieldName, sub.Item.FieldId)
	}

	eja.SqlQuery = sqlQuery + sqlLinks + db.SearchQueryOrderAndLimit(sqlOrder, eja.SearchLimit, eja.SearchOffset)
	eja.SearchCount = db.SearchCount(sqlQuery+sqlLinks, eja.SqlQueryArgs)
	eja.SearchLast = min(eja.SearchOffset+eja.SearchLimit, eja.SearchCount)

	db.SessionPut(eja.Owner, "SearchLimit", db.String(eja.SearchLimit))
	db.SessionPut(eja.Owner, "SearchOffset", db.String(eja.SearchOffset))
	eja.Id = 0
	return eja
}

func wrapUp(eja Api, db DbSession, sessionSave bool) Api {
	if eja.Linking {
		db.SessionPut(eja.Owner, "Link", db.String(eja.Link.ModuleId), "ModuleId")
		db.SessionPut(eja.Owner, "Link", db.String(eja.Link.FieldId), "FieldId")
		db.SessionPut(eja.Owner, "Link", eja.Link.Label, "Label")
		eja.SearchLinks = db.SearchLinks(eja.Owner, eja.Link.ModuleId, eja.Link.FieldId, eja.ModuleId)
	}

	eja.ModuleLabel = db.Translate(eja.ModuleName, eja.Owner)

	if eja.ActionType == "List" {
		eja.SearchRows, eja.SearchCols, eja.SearchLabels, _ = db.SearchMatrix(eja.Owner, eja.ModuleId, eja.SqlQuery, eja.SqlQueryArgs)
	} else if eja.Id > 0 {
		eja.ActionType = "Edit"
		eja.Links = db.ModuleLinks(eja.Owner, eja.ModuleId)
		eja.SubModules = db.SubModules(eja.Owner, eja.ModuleId)
		db.SessionPut(eja.Owner, "ejaId", db.String(eja.Id))
	} else {
		eja.ActionType = "Search"
		db.SessionCleanSearch(eja.Owner)
	}

	eja.Commands, _ = db.Commands(eja.Owner, eja.ModuleId, eja.ActionType)
	eja.Fields, _ = db.Fields(eja.Owner, eja.ModuleId, eja.ActionType, eja.Values)
	eja.Path = db.ModulePath(eja.Owner, eja.ModuleId)
	eja.Tree = db.ModuleTree(eja.Owner, eja.ModuleId, eja.Path)

	if plugin, ok := Plugins[eja.ModuleName]; ok {
		eja = plugin(eja, db)
	}

	if eja.Owner > 0 && !sessionSave {
		db.SessionReset(eja.Owner)
	}
	return eja
}

type ActiveSubModule struct {
	Item  SubModulePathItem
	Found bool
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
