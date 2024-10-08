// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/eja/tibula/sys"
)

func Set() TypeApi {
	return TypeApi{
		Language:           sys.Options.Language,
		DefaultSearchLimit: 15,
		DefaultSearchOrder: "ejaId DESC",
		Values:             make(map[string]string),
		SearchOrder:        make(map[string]string),
		Link:               TypeDbLink{},
	}
}

func Run(eja TypeApi, sessionSave bool) (result TypeApi, err error) {
	var user map[string]string

	db := DbSession()

	//open db connection
	if err = db.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort); err != nil {
		return
	}

	//login and logout
	if eja.Action == "login" {
		if eja.Values["username"] != "" && eja.Values["password"] != "" {
			user = db.UserGetAllByUserAndPass(eja.Values["username"], eja.Values["password"])
			if len(user) > 0 {
				eja.Session = db.SessionInit(db.Number(user["ejaId"]))
			}
		}
	}
	if eja.Values["googleSsoToken"] != "" {
		ssoUsername := googleSsoEmail(eja.Values["googleSsoToken"])
		if ssoUsername != "" {
			user = db.UserGetAllByUsername(ssoUsername)
		}
		if len(user) > 0 {
			eja.Session = db.SessionInit(db.Number(user["ejaId"]))
		}
	}
	if eja.Session != "" {
		if len(user) == 0 {
			user = db.UserGetAllBySession(eja.Session)
		}
		if len(user) > 0 {
			eja.Owner = db.Number(user["ejaId"])
			if user["ejaLanguage"] != "" {
				eja.Language = user["ejaLanguage"]
			}
		}
		if eja.ModuleId == 0 && eja.ModuleName != "" {
			eja.ModuleId = db.ModuleGetIdByName(eja.ModuleName)
		}
		if eja.ModuleId == 0 {
			eja.ModuleId = db.Number(user["defaultModuleId"])
			eja.ModuleName = db.ModuleGetNameById(eja.ModuleId)
		}
	}
	if eja.Action == "logout" && eja.Owner > 0 {
		db.SessionReset(eja.Owner)
		eja.Session = ""
		eja.Owner = 0
	}

	if eja.Owner == 0 {
		var error = "ejaNotAuthorized"
		if len(eja.Values) > 0 {
			alert(&eja.Alert, db.Translate(error))
		}
		return eja, errors.New(error)
	} else {
		// set module id and name
		if eja.ModuleId == 0 {
			eja.ModuleId = db.ModuleGetIdByName("eja")
		}
		if eja.ModuleName == "" {
			eja.ModuleName = db.ModuleGetNameById(eja.ModuleId)
		}

		//check if the operation is permitted
		eja.Commands, _ = db.Commands(eja.Owner, eja.ModuleId, "")
		if eja.Action != "" && eja.Action != "login" && !db.CommandExists(eja.Commands, eja.Action) {
			alert(&eja.Alert, db.Translate("ejaNotPermitted", eja.Owner))
		} else {
			//session
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
			//link
			linkingField := ""
			if eja.Link.ModuleId > 0 && eja.Link.FieldId > 0 && eja.Link.Label != "" {
				eja.Linking = true
				eja.Link.ModuleLabel = db.Translate(db.ModuleGetNameById(eja.Link.ModuleId), eja.Owner)
				linkingField = db.ModuleLinksFieldName(eja.ModuleId, eja.Link.ModuleId)
				if linkingField != "" && eja.Action != "search" {
					eja.Values[linkingField] = db.String(eja.Link.FieldId)
				}
				if eja.ModuleId == eja.Link.ModuleId && eja.Id == eja.Link.FieldId && eja.Id > 0 {
					//reset
					db.SessionCleanLink(eja.Owner)
					db.SessionCleanSearch(eja.Owner)
					eja.Link = TypeDbLink{}
					eja.Action = "edit"
					eja.Linking = false
				}
				for _, fieldId := range eja.IdList {
					if eja.Action == "link" {
						db.LinkDel(eja.Owner, eja.ModuleId, db.Number(fieldId), eja.Link.ModuleId, eja.Link.FieldId)
						db.LinkAdd(eja.Owner, eja.ModuleId, db.Number(fieldId), eja.Link.ModuleId, eja.Link.FieldId)
					}
					if eja.Action == "unlink" {
						db.LinkDel(eja.Owner, eja.ModuleId, db.Number(fieldId), eja.Link.ModuleId, eja.Link.FieldId)
					}
				}
				if eja.Action == "link" || eja.Action == "unlink" {
					eja.IdList = []int64{}
					eja.ActionType = "List"
				}
			}
			//edit
			if eja.Id > 0 && eja.Action == "edit" {
				eja.Values = db.TableGetAllById(eja.ModuleName, eja.Id)
			}
			//new or copy
			if eja.Action == "new" || eja.Action == "copy" {
				copyId := eja.Id
				eja.Id, _ = db.New(eja.Owner, eja.ModuleId)
				if eja.Id < 1 {
					alert(&eja.Alert, db.Translate("ejaActionNewError", eja.Owner))
				} else {
					db.LinkCopy(eja.Owner, eja.Id, eja.ModuleId, copyId)
				}
			}
			//save, copy, new
			if len(eja.Values) > 0 && (eja.Action == "save" || eja.Action == "copy" || eja.Action == "new") {
				if eja.ModuleName == "ejaModules" {
					if db.Number(eja.Values["sqlCreated"]) > 0 {
						if err := db.TableAdd(eja.Values["name"]); err != nil {
							alert(&eja.Alert, db.Translate("ejaSqlModuleNotCreated", eja.Owner))
						} else {
							info(&eja.Info, db.Translate("ejaSqlModuleCreated", eja.Owner))
						}
					}
					if eja.Action == "save" && db.PermissionCount(eja.Id) == 0 {
						if db.Number(eja.Values["sqlCreated"]) > 0 {
							db.PermissionAddDefault(eja.Owner, eja.Id)
							info(&eja.Info, db.Translate("ejaModulePermissionsAddDefault", eja.Owner))
						} else {
							db.PermissionAdd(eja.Owner, eja.Id, "logout")
							info(&eja.Info, db.Translate("ejaModulePermissionAdd", eja.Owner))
						}
						db.UserPermissionCopy(eja.Owner, eja.Id)
					}
				}
				if eja.ModuleName == "ejaFields" {
					moduleName := db.ModuleGetNameById(db.Number(eja.Values["ejaModuleId"]))
					if err := db.FieldAdd(moduleName, eja.Values["name"], eja.Values["type"]); err == nil {
						info(&eja.Info, db.Translate("ejaSqlFieldCreated", eja.Owner))
					} else {
						alert(&eja.Alert, db.Translate("ejaSqlFieldNotCreated", eja.Owner))
					}
				}
				if eja.Id < 1 {
					id, err := db.New(eja.Owner, eja.ModuleId)
					if err == nil {
						eja.Id = id
					}
				}
				if eja.Id < 1 {
					alert(&eja.Alert, db.Translate("ejaErrorEditId", eja.Owner))
				} else {
					for key, val := range eja.Values {
						var value interface{}
						fieldType := db.FieldTypeGet(eja.ModuleId, key)
						switch fieldType {
						case "password":
							if len(val) != 64 {
								value = db.Sha256(val)
							}
						case "boolean", "integer":
							value = db.Number(val)
						case "decimal":
							value = db.Float(val)
						default:
							value = db.String(val)
						}
						if key == "ejaOwner" && db.Number(val) < 1 {
							eja.Values["ejaOwner"] = db.String(eja.Owner)
						} else {
							db.Put(eja.Owner, eja.ModuleId, eja.Id, key, value)
						}
					}
					values, err := db.Get(eja.Owner, eja.ModuleId, eja.Id)
					if err == nil {
						eja.Values = values
					}
				}
			}
			//delete
			if eja.Action == "delete" {
				if len(eja.IdList) == 0 && eja.Id > 0 {
					eja.IdList = append(eja.IdList, eja.Id)
				}
				for _, val := range eja.IdList {
					err := db.Del(eja.Owner, eja.ModuleId, val)
					if eja.ModuleName == "ejaModules" {
						if err == nil {
							info(&eja.Info, db.Translate("ejaSqlModuleDeleteTrue", eja.Owner))
						} else {
							alert(&eja.Alert, db.Translate("ejaSqlModuleDeleteFalse", eja.Owner))
						}
					}
				}
				eja.ActionType = "List"
			}

			//list
			if eja.Action == "list" && eja.SqlQuery64 == "" {
				eja.Values = make(map[string]string)
				eja.Action = ""
				eja.Id = 0
			}

			//search
			if eja.Action == "search" || eja.Action == "previous" || eja.Action == "next" || eja.Action == "list" || eja.ActionType == "List" {
				var sqlQuery string
				var sqlArgs []interface{}
				var sqlOrderFields = db.FieldNameList(eja.ModuleId, "List")
				var sqlOrder string
				var sqlLinks string
				var err error
				eja.ActionType = "List"
				moduleDefault := db.TableGetAllById("ejaModules", eja.ModuleId)

				//limit
				if eja.SearchLimit < 1 {
					eja.SearchLimit = db.Number(moduleDefault["searchLimit"])
				}
				if eja.SearchLimit < 1 {
					eja.SearchLimit = eja.DefaultSearchLimit
				}

				//previous and next
				if eja.Action == "previous" && eja.SearchOffset >= eja.SearchLimit {
					eja.SearchOffset -= eja.SearchLimit
				}
				if eja.Action == "next" {
					eja.SearchOffset += eja.SearchLimit
				}

				//query
				if eja.SqlQuery64 != "" {
					sqlQueryBytes, err := base64.StdEncoding.DecodeString(eja.SqlQuery64)
					if err == nil {
						sqlQuery = db.String(sqlQueryBytes)
					}
				} else {
					sqlQuery, sqlArgs, err = db.SearchQuery(eja.Owner, eja.ModuleName, eja.Values)
					if err == nil {
						eja.SqlQuery64 = base64.StdEncoding.EncodeToString([]byte(sqlQuery))
						eja.SqlQueryArgs = sqlArgs
						db.SessionPut(eja.Owner, "SqlQuery64", eja.SqlQuery64)
						for _, row := range sqlArgs {
							db.SessionPut(eja.Owner, "SqlQueryArgs", db.String(row))
						}
					}
				}

				//order
				for _, key := range sqlOrderFields {
					value := eja.SearchOrder[key]
					if value != "" && (value == "ASC" || value == "DESC") {
						db.SessionPut(eja.Owner, "SearchOrder", value, key)
						if sqlOrder == "" {
							sqlOrder += key + " " + value
						} else {
							sqlOrder += "," + key + " " + value
						}
					}
				}
				if sqlOrder == "" && moduleDefault["sortList"] != "" {
					sqlOrder = moduleDefault["sortList"] + " ASC"
				}
				if sqlOrder == "" {
					sqlOrder = eja.DefaultSearchOrder
				}

				//link
				if linkingField != "" {
					sqlLinks = fmt.Sprintf(" AND %s=? ", linkingField)
					eja.SqlQueryArgs = append(eja.SqlQueryArgs, db.String(eja.Link.FieldId))
				} else if eja.SearchLink {
					sqlLinks = db.SearchQueryLinks(eja.Owner, eja.Link.ModuleId, eja.Link.FieldId, eja.ModuleId)
				}

				eja.SqlQuery = sqlQuery + sqlLinks + db.SearchQueryOrderAndLimit(sqlOrder, eja.SearchLimit, eja.SearchOffset)
				eja.SearchCount = db.SearchCount(sqlQuery+sqlLinks, eja.SqlQueryArgs)
				eja.SearchLast = eja.SearchOffset + eja.SearchLimit
				if eja.SearchLast > eja.SearchCount {
					eja.SearchLast = eja.SearchCount
				}
				db.SessionPut(eja.Owner, "SearchLimit", db.String(eja.SearchLimit))
				db.SessionPut(eja.Owner, "SearchOffset", db.String(eja.SearchOffset))
				eja.Id = 0
			}

			//linking last step
			if eja.Linking {
				db.SessionPut(eja.Owner, "Link", db.String(eja.Link.ModuleId), "ModuleId")
				db.SessionPut(eja.Owner, "Link", db.String(eja.Link.FieldId), "FieldId")
				db.SessionPut(eja.Owner, "Link", eja.Link.Label, "Label")
				eja.SearchLinks = db.SearchLinks(eja.Owner, eja.Link.ModuleId, eja.Link.FieldId, eja.ModuleId)
				if linkingField != "" {
					eja.Linking = false
				}
			}
		}

		eja.ModuleLabel = db.Translate(eja.ModuleName, eja.Owner)

		if eja.ActionType == "List" {
			eja.SearchRows, eja.SearchCols, eja.SearchLabels, _ = db.SearchMatrix(eja.Owner, eja.ModuleId, eja.SqlQuery, eja.SqlQueryArgs)
		} else {
			if eja.Id > 0 {
				eja.ActionType = "Edit"
				eja.Links = db.ModuleLinks(eja.Owner, eja.ModuleId)
				db.SessionPut(eja.Owner, "ejaId", db.String(eja.Id))
			} else {
				eja.ActionType = "Search"
				db.SessionCleanSearch(eja.Owner)
			}
		}
		eja.Commands, _ = db.Commands(eja.Owner, eja.ModuleId, eja.ActionType)
		eja.Fields, _ = db.Fields(eja.Owner, eja.ModuleId, eja.ActionType, eja.Values)
		eja.Path = db.ModulePath(eja.Owner, eja.ModuleId)
		eja.Tree = db.ModuleTree(eja.Owner, eja.ModuleId, eja.Path)
	}

	if Plugins[eja.ModuleName] != nil {
		eja = Plugins[eja.ModuleName](eja, db)
	}

	if eja.Owner > 0 && !sessionSave {
		db.SessionReset(eja.Owner)
	}

	db.Close()
	return eja, nil
}
