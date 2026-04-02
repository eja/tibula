// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func (a *Api) info(value string) {
	a.Info = append(a.Info, value)
	slog.Debug(value, "gui", "info")
}

func (a *Api) alert(value string) {
	a.Alert = append(a.Alert, value)
	slog.Debug(value, "gui", "alert")
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
