// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/sys"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets
var assets embed.FS

func Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", run)
	if sys.Options.WebPath != "" {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(sys.Options.WebPath, "static")))))
	} else {
		staticFs, err := fs.Sub(assets, "assets/static")
		if err != nil {
			return err
		}
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFs))))
	}

	address := fmt.Sprintf("%s:%d", sys.Options.WebHost, sys.Options.WebPort)

	if sys.Options.WebTlsPrivate != "" && sys.Options.WebTlsPublic != "" {
		if _, err := os.Stat(sys.Options.WebTlsPrivate); err != nil {
			return errors.New("cannot open private certificate")
		} else {
			if _, err := os.Stat(sys.Options.WebTlsPublic); err != nil {
				return errors.New("cannot open public certificate")
			} else {
				log.Printf("Starting server on https://%s\n", address)
				if err := http.ListenAndServeTLS(address, sys.Options.WebTlsPublic, sys.Options.WebTlsPrivate, mux); err != nil {
					return err
				}
			}
		}
	} else {
		log.Printf("Starting server on http://%s\n", address)
		if err := http.ListenAndServe(address, mux); err != nil {
			return err
		}
	}

	return nil
}

func run(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Tibula/"+sys.Version)
	templateFile := "Login.html"
	var err error
	var user map[string]string
	eja := TypeEja{
		Language:           sys.Options.Language,
		DefaultSearchLimit: 15,
		DefaultSearchOrder: "ejaLog DESC",
		Values:             make(map[string]string),
		SearchOrder:        make(map[string]string),
		Link:               db.TypeLink{},
	}

	r.PostFormValue("")
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//open db connection
	db.LogLevel = sys.Options.LogLevel
	if err := db.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//process GET parameters
	for key, values := range r.URL.Query() {
		value := values[0]
		switch key {
		case "ejaSession":
			eja.Session = value
		case "ejaId":
			eja.Id = db.Number(value)
		case "ejaModuleId":
			eja.ModuleId = db.Number(value)
			eja.ModuleName = db.ModuleGetNameById(eja.ModuleId)
		case "ejaLink":
			parts := strings.Split(value, ".")
			if len(parts) == 3 {
				eja.Link = db.TypeLink{
					ModuleId: db.Number(parts[0]),
					FieldId:  db.Number(parts[1]),
					Label:    parts[2],
				}
				eja.Action = "search"
				eja.SearchLink = true
				eja.SearchLinkClean = true
			}
		}
	}

	//process POST parameters
	for key, values := range r.Form {
		value := values[0]
		switch key {
		case "ejaId":
			eja.Id = db.Number(value)
		case "ejaModuleId":
			if eja.ModuleName == "" {
				eja.ModuleId = db.Number(value)
			}
		case "ejaModuleName":
			eja.ModuleName = value
			if eja.ModuleId == 0 {
				eja.ModuleId = db.ModuleGetIdByName(value)
			}
		case "ejaAction":
			eja.Action = value
		case "ejaSession":
			eja.Session = value
		case "ejaSearchLink":
			if db.Number(value) > 0 {
				eja.SearchLink = true
			} else {
				eja.SearchLink = false
			}
		}
	}
	for key, value := range r.Form {
		if strings.HasPrefix(key, "ejaValues") {
			eja.Values[arrayKeyNameExtract(key)] = value[0]
		}
		if strings.HasPrefix(key, "ejaSearchOrder") {
			order := strings.ToUpper(value[0])
			if order == "ASC" || order == "DESC" || order == "NONE" {
				eja.SearchOrder[arrayKeyNameExtract(key)] = order
			}
		}
		if strings.HasPrefix(key, "ejaIdList") {
			id := db.Number(arrayKeyNameExtract(key))
			if id > 0 {
				if len(eja.IdList) == 0 {
					eja.Id = id
				}
				eja.IdList = append(eja.IdList, id)
			}
		}
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
		if len(eja.Values) > 0 {
			alert(&eja.Alert, db.Translate("ejaNotAuthorized"))
		}
	} else {
		// set module id and name
		if eja.ModuleId == 0 {
			eja.ModuleId = db.ModuleGetIdByName("eja")
		}
		if eja.ModuleName == "" {
			eja.ModuleName = db.ModuleGetNameById(eja.ModuleId)
		}

		//check if the operation is permitted
		eja.Commands, err = db.Commands(eja.Owner, eja.ModuleId, "")
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
			if eja.Link.ModuleId > 0 && eja.Link.FieldId > 0 && eja.Link.Label != "" {
				eja.Link.ModuleLabel = db.Translate(db.ModuleGetNameById(eja.Link.ModuleId), eja.Owner)
				eja.Linking = true
				if eja.ModuleId == eja.Link.ModuleId && eja.Id == eja.Link.FieldId && eja.Id > 0 {
					db.SessionCleanLink(eja.Owner)
					db.SessionCleanSearch(eja.Owner)
					eja.Link = db.TypeLink{}
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
					eja.Action = "search"
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
						fieldType := db.FieldTypeGet(eja.ModuleId, key)
						switch fieldType {
						case "password":
							if len(val) != 64 {
								val = db.Sha256(val)
							}
						}
						if key == "ejaOwner" && db.Number(val) < 1 {
							eja.Values["ejaOwner"] = db.String(eja.Owner)
						} else {
							db.Put(eja.Owner, eja.ModuleId, eja.Id, key, val)
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
				eja.Action = "search"
			}
			//search
			if eja.Action == "search" || eja.Action == "previous" || eja.Action == "next" || eja.Action == "list" {
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
				if eja.SearchLink {
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
		}

		//linking
		if eja.Linking {
			db.SessionPut(eja.Owner, "Link", db.String(eja.Link.ModuleId), "ModuleId")
			db.SessionPut(eja.Owner, "Link", db.String(eja.Link.FieldId), "FieldId")
			db.SessionPut(eja.Owner, "Link", eja.Link.Label, "Label")
			eja.SearchLinks = db.SearchLinks(eja.Owner, eja.Link.ModuleId, eja.Link.FieldId, eja.ModuleId)
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
			eja.Fields, _ = db.Fields(eja.Owner, eja.ModuleId, eja.ActionType, eja.Values)
		}
		templateFile = eja.ActionType + ".html"
		eja.Commands, _ = db.Commands(eja.Owner, eja.ModuleId, eja.ActionType)
		eja.Path = db.ModulePath(eja.Owner, eja.ModuleId)
		eja.Tree = db.ModuleTree(eja.Owner, eja.ModuleId, eja.Path)
	}

	db.Close()

	if r.Header.Get("Content-Type") == "application/json" {
		jsonData, err := json.Marshal(eja)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonData); err != nil {
			log.Fatal("cannot return json data")
		}

	} else {
		var tpl *template.Template
		if sys.Options.WebPath != "" {
			tpl, err = template.ParseGlob(filepath.Join(sys.Options.WebPath, "templates", "*.html"))
		} else {
			tpl, err = template.ParseFS(assets, "assets/templates/*.html")
		}
		if err != nil || tpl.ExecuteTemplate(w, templateFile, eja) != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if err != nil {
		log.Println(err)
	}
}
