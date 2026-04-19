// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/eja/tibula/api"
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/sys"
)

type loginRateLimit struct {
	failures int
	lockout  time.Time
}

var loginTracker sync.Map

const maxLoginFailures = 5
const loginLockoutDuration = 5 * time.Minute

func updateLoginTracker(ip string, action string, err error) {
	if action != "login" {
		return
	}

	if err != nil && err.Error() == "ejaNotAuthorized" {
		val, _ := loginTracker.LoadOrStore(ip, loginRateLimit{})
		limit := val.(loginRateLimit)
		limit.failures++
		if limit.failures >= maxLoginFailures {
			limit.lockout = time.Now().Add(loginLockoutDuration)
		}
		loginTracker.Store(ip, limit)
	} else if err == nil {
		loginTracker.Delete(ip)
	}
}

func Core(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", sys.Label+"/"+sys.Version)

	clientIP := getClientIP(r)

	if val, ok := loginTracker.Load(clientIP); ok {
		limit := val.(loginRateLimit)
		if limit.failures >= maxLoginFailures {
			if time.Now().Before(limit.lockout) {
				slog.Warn("IP locked out due to brute-force protection", "ip", clientIP)
				http.Error(w, "Too many failed login attempts. Try again later.", http.StatusTooManyRequests)
				return
			} else {
				loginTracker.Delete(clientIP)
			}
		}
	}

	templateFile := "Login.html"
	var err error

	if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
		eja := api.Set()
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&eja)
		if err != nil {
			msg := "JSON structure is not valid"
			slog.Error(msg, "address", r.RemoteAddr, "error", err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		eja, err = api.Run(eja, false)
		updateLoginTracker(clientIP, eja.Action, err)
		if err != nil {
			if err.Error() == "ejaNotAuthorized" {
				slog.Warn("API login problem", "address", r.RemoteAddr, "error", err)
				http.Error(w, "Unauthorized: Access Denied", http.StatusUnauthorized)
			} else {
				slog.Error("API process error", "address", r.RemoteAddr, "error", err)
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
				slog.Error("cannot return json data", "error", err)
			}
		}

	} else {
		eja := api.Set()
		eja.RemoteIP = getClientIP(r)
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
			case "ejaSubModulePath":
				eja.SubModulePath = subModulePathExtract(value)
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
			case "ejaSubModulePath":
				eja.SubModulePath = subModulePathExtract(value)
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

		if len(r.Form) == 0 {
			err = nil
		} else {
			eja, err = api.Run(eja, true)
			updateLoginTracker(clientIP, eja.Action, err)
			if err != nil {
				if err.Error() == "ejaNotAuthorized" {
					slog.Warn("API login problem", "address", r.RemoteAddr, "error", err)
				} else {
					slog.Error("API process error", "address", r.RemoteAddr, "error", err)
				}
			} else {
				templateFile = eja.ActionType + ".html"
			}
		}

		if sys.Options.GoogleSsoId != "" {
			eja.GoogleSsoId = sys.Options.GoogleSsoId
		}

		var tpl *template.Template
		templateFunctions := template.FuncMap{
			"csvContains": csvContains,
			"safe":        func(s string) template.HTML { return template.HTML(s) },
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
		slog.Error("API process error", "address", r.RemoteAddr, "error", err)
	}
}
