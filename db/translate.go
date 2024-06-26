// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"github.com/eja/tibula/log"
)

// Translate retrieves the translation for the specified word based on the current session's language and module context.
// If a translation is not found, it returns a placeholder string indicating that the translation is missing.
func (session *TypeSession) Translate(value string, user ...int64) string {
	var userId int64
	var result string
	if len(user) > 0 {
		userId = user[0]
	}
	if userId > 0 {
		result, _ = session.Value(`
			SELECT translation
			FROM ejaTranslations
			WHERE word = ?
 	   		AND ejaLanguage = (SELECT ejaLanguage FROM ejaUsers WHERE ejaUsers.ejaId = ?)
 	   		AND (
        	ejaModuleId = 0
        	OR ejaModuleId = ''
        	OR ejaModuleId = (
        	SELECT value FROM ejaSessions WHERE ejaSessions.name = 'ejaModuleId' AND ejaSessions.ejaOwner = ?)
    		)
			ORDER BY ejaModuleId DESC
			LIMIT 1
			`, value, userId, userId)
	} else {
		result, _ = session.Value("SELECT translation FROM ejaTranslations WHERE word=? AND (ejaLanguage=0 OR ejaLanguage='') LIMIT 1", value)
	}
	if result == "" {
		if log.Level >= log.LevelDebug {
			result = "{" + value + "}"
		} else {
			result = value
		}
	}
	return result
}
