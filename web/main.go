// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eja/tibula/core"
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
	address := fmt.Sprintf("%s:%d", sys.Options.WebHost, sys.Options.WebPort)

	mux.HandleFunc("/", root)

	if sys.Options.WebPath != "" {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(sys.Options.WebPath, "static")))))
	} else {
		staticFs, err := fs.Sub(assets, "assets/static")
		if err != nil {
			return err
		}
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFs))))
	}

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

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Tibula/"+sys.Version)
	templateFile := "Login.html"
	var err error
	eja := core.Init()

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
			eja.Id = db.Number(value)
		case "ejaModuleId":
			eja.ModuleId = db.Number(value)
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
			eja.ModuleId = db.Number(value)
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

	eja, err = core.Run(eja)
	if err != nil {
		log.Println(r.RemoteAddr, err)
	} else {
		templateFile = eja.ActionType + ".html"
	}

	if r.Header.Get("Content-Type") == "application/json" {
		jsonData, err := json.Marshal(eja)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err = w.Write(jsonData); err != nil {
			log.Fatal(r.RemoteAddr, "cannot return json data")
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
		log.Println(r.RemoteAddr, err)
	}
}
