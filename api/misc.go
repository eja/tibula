// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/eja/tibula/log"
	"github.com/eja/tibula/sys"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func (a *Api) info(value string) {
	a.Info = append(a.Info, value)
	if sys.Options.LogLevel > 3 {
		log.Trace(tag, "[info]", value)
	}
}

func (a *Api) alert(value string) {
	a.Alert = append(a.Alert, value)
	if sys.Options.LogLevel > 3 {
		log.Trace(tag, "[alert]", value)
	}
}

func googleSsoEmail(token string) string {
	resp, err := httpClient.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + token)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			return result.Email
		}
	}
	return ""
}
