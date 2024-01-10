// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"math/rand"
	"strconv"
)

// SessionInit generates a new session for the specified user and updates the database.
func SessionInit(userId int64) string {
	session := Sha256(strconv.Itoa(rand.Int()) + strconv.Itoa(rand.Int()))
	Run("UPDATE ejaUsers SET ejaSession=? WHERE ejaId=?", session, userId)
	Run("DELETE FROM ejaSessions WHERE ejaOwner=?", userId)
	return session
}

// SessionLoad loads session data for a specified user and module from the database.
func SessionLoad(userId int64, moduleId int64) (TypeRows, error) {
	if DbEngine == "mysql" {
		Run("SET @ejaOwner = " + String(userId))
	}
	TableAdd("ejaSession", true)
	FieldAdd("ejaSession", "name", "text")
	FieldAdd("ejaSession", "value", "text")
	FieldAdd("ejaSession", "sub", "text")
	Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name in ('ejaId','ejaOwners')", userId)
	Run("INSERT INTO ejaSession SELECT * FROM ejaSessions WHERE ejaOwner=?", userId)
	user := UserGetAllById(userId)
	SessionPut(userId, "ejaModuleId", String(moduleId))
	SessionPut(userId, "ejaModuleName", ModuleGetNameById(moduleId))
	SessionPut(userId, "ejaOwner", String(user["ejaId"]))
	SessionPut(userId, "ejaLanguage", user["ejaLanguage"])
	for _, val := range Owners(userId, moduleId) {
		SessionPut(userId, "ejaOwners", String(val), String(val))
	}
	return Rows("SELECT * FROM ejaSession ORDER BY ejaLog ASC, ejaId ASC")
}

// SessionPut stores a session variable for a specified user in the database.
func SessionPut(userId int64, name string, value string, subName ...string) (err error) {
	var sub string
	if len(subName) > 0 {
		sub = subName[0]
	}
	if _, err = Run("DELETE FROM ejaSession WHERE ejaOwner=? AND name=? AND sub=?", userId, name, sub); err != nil {
		return
	}
	if _, err = Run("INSERT INTO ejaSession (ejaId, ejaOwner, ejaLog, name, value, sub) VALUES (NULL,?,?,?,?,?)", userId, Now(), name, value, sub); err != nil {
		return
	}
	if _, err = Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name=? AND sub=?", userId, name, sub); err != nil {
		return
	}
	if _, err = Run("INSERT INTO ejaSessions (ejaId, ejaOwner, ejaLog, name, value, sub) VALUES (NULL,?,?,?,?,?)", userId, Now(), name, value, sub); err != nil {
		return err
	}
	return
}

// SessionCleanLink removes specific session variables related to links for a specified user.
func SessionCleanLink(userId int64) error {
	_, err := Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name in ('Link','SqlQuery64','SqlQueryArgs','SearchLimit','SearchOffset','SearchOrder')", userId)
	return err
}

// SessionCleanSearch removes specific session variables related to searches for a specified user.
func SessionCleanSearch(userId int64) error {
	_, err := Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name in ('SqlQuery64','SqlQueryArgs','SearchLimit','SearchOffset','SearchOrder')", userId)
	return err
}

// SessionReset removes all session variables for a specified user from the database.
func SessionReset(userId int64) error {
	_, err := Run("DELETE FROM ejaSessions WHERE ejaOwner=?", userId)
	return err
}
