// Copyright (C) 2007-2025 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/eja/tibula/api"
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/log"
	"github.com/eja/tibula/sys"
)

func Core(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", sys.Name+"/"+sys.Version)
	templateFile := "Login.html"
	var err error

	if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
		eja := api.Set()
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&eja)
		if err != nil {
			log.Error("[web]", r.RemoteAddr, err)
			http.Error(w, "JSON structure is not valid", http.StatusBadRequest)
			return
		}
		eja, err = api.Run(eja, false)
		if err != nil {
			log.Error("[web]", r.RemoteAddr, err)
			if err.Error() == "ejaNotAuthorized" {
				http.Error(w, "Unauthorized: Access Denied", http.StatusUnauthorized)
			} else {
				http.Error(w, "API process error", http.StatusInternalServerError)
			}
		} else {
			eja.SqlQuery = ""
			eja.SqlQuery64 = ""
			eja.SqlQueryArgs = nil
			eja.DefaultSearchLimit = 0
			eja.DefaultSearchOrder = ""
			jsonData, err := json.Marshal(eja)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err = w.Write(jsonData); err != nil {
				log.Fatal(r.RemoteAddr, "cannot return json data")
			}
		}

	} else {
		eja := api.Set()
		r.PostFormValue("")
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		//process GET parameters
		for key, values := range r.URL.Query() {
			value := values[0]
			switch key {
			case "ejaSession":
				eja.Session = value
			case "ejaId":
				eja.Id = sys.Number(value)
			case "ejaModuleId":
				eja.ModuleId = sys.Number(value)
			case "ejaLink":
				parts := strings.Split(value, ".")
				if len(parts) == 3 {
					eja.Link = db.TypeLink{
						ModuleId: sys.Number(parts[0]),
						FieldId:  sys.Number(parts[1]),
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
				eja.Id = sys.Number(value)
			case "ejaModuleId":
				eja.ModuleId = sys.Number(value)
			case "ejaAction":
				eja.Action = value
			case "ejaSession":
				eja.Session = value
			case "ejaSearchLink":
				if sys.Number(value) > 0 {
					eja.SearchLink = true
				} else {
					eja.SearchLink = false
				}
			}
		}
		for key, value := range r.Form {
			if strings.HasPrefix(key, "ejaValues") {
				if len(value) > 1 {
					eja.Values[arrayKeyNameExtract(key)] = arrayToCsvQuoted(value)
				} else {
					eja.Values[arrayKeyNameExtract(key)] = value[0]
				}
			}
			if strings.HasPrefix(key, "ejaSearchOrder") {
				order := strings.ToUpper(value[0])
				if order == "ASC" || order == "DESC" || order == "NONE" {
					eja.SearchOrder[arrayKeyNameExtract(key)] = order
				}
			}
			if strings.HasPrefix(key, "ejaIdList") {
				id := sys.Number(arrayKeyNameExtract(key))
				if id > 0 {
					if len(eja.IdList) == 0 {
						eja.Id = id
					}
					eja.IdList = append(eja.IdList, id)
				}
			}
		}

		eja, err = api.Run(eja, true)
		if err != nil {
			log.Error("[web]", r.RemoteAddr, err)
		} else {
			templateFile = eja.ActionType + ".html"
		}

		if sys.Options.GoogleSsoId != "" {
			eja.GoogleSsoId = sys.Options.GoogleSsoId
		}

		var tpl *template.Template
		templateFunctions := template.FuncMap{
			"csvContains": csvContains,
		}
		if sys.Options.WebPath != "" {
			tpl, err = template.New("").Funcs(templateFunctions).ParseGlob(filepath.Join(sys.Options.WebPath, "templates", "*.html"))
		} else {
			tpl, err = template.New("").Funcs(templateFunctions).ParseFS(assets, "assets/templates/*.html")
		}
		if err != nil || tpl.ExecuteTemplate(w, templateFile, eja) != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if err != nil {
		log.Error("[web]", r.RemoteAddr, err)
	}
}
