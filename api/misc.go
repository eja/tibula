// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eja/tibula/log"
	"github.com/eja/tibula/sys"
)

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

func googleSsoEmail(token string) (email string) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + token)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			email, _ = result["email"].(string)
		}
	}
	return
}
