// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eja/tibula/log"
	"github.com/eja/tibula/sys"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func info(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	*array = append(*array, row)
	if sys.Options.LogLevel > 3 {
		log.Trace("[api] [info]", row)
	}
}

func alert(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	*array = append(*array, row)
	if sys.Options.LogLevel > 3 {
		log.Trace("[api] [alert]", row)
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
