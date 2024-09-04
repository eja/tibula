// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	randCrypto "crypto/rand"
	"encoding/binary"
	"fmt"
	randMath "math/rand"
	"time"
)

// SessionInit generates a new session for the specified user and updates the database.
func (session *TypeSession) SessionInit(userId int64) string {
	var seed int64
	maxValue := 1000 * 1000 * 1000
	randBytes := make([]byte, 8)
	n, err := randCrypto.Read(randBytes)
	if err != nil || n != 8 {
		seed = time.Now().Unix()
	} else {
		seed = int64(binary.LittleEndian.Uint64(randBytes))
	}
	randMath.Seed(seed)

	sessionHash := session.Sha256(fmt.Sprintf("%d%d", randMath.Intn(maxValue), randMath.Intn(maxValue)))
	session.Run("UPDATE ejaUsers SET ejaSession=? WHERE ejaId=?", sessionHash, userId)
	session.Run("DELETE FROM ejaSessions WHERE ejaOwner=?", userId)
	return sessionHash
}

// SessionLoad loads session data for a specified user and module from the database.
func (session *TypeSession) SessionLoad(userId int64, moduleId int64) (TypeRows, error) {
	if session.Engine == "mysql" {
		session.Run("SET @ejaOwner = " + session.String(userId))
	}
	session.TableAdd("ejaSession", true)
	session.FieldAdd("ejaSession", "name", "text")
	session.FieldAdd("ejaSession", "value", "text")
	session.FieldAdd("ejaSession", "sub", "text")
	session.Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name in ('ejaId','ejaOwners')", userId)
	session.Run("INSERT INTO ejaSession SELECT * FROM ejaSessions WHERE ejaOwner=?", userId)
	user := session.UserGetAllById(userId)
	session.SessionPut(userId, "ejaModuleId", session.String(moduleId))
	session.SessionPut(userId, "ejaModuleName", session.ModuleGetNameById(moduleId))
	session.SessionPut(userId, "ejaOwner", session.String(user["ejaId"]))
	session.SessionPut(userId, "ejaLanguage", user["ejaLanguage"])
	for _, val := range session.Owners(userId, moduleId) {
		session.SessionPut(userId, "ejaOwners", session.String(val), session.String(val))
	}
	return session.Rows("SELECT * FROM ejaSession ORDER BY ejaLog ASC, ejaId ASC")
}

// SessionPut stores a session variable for a specified user in the database.
func (session *TypeSession) SessionPut(userId int64, name string, value string, subName ...string) (err error) {
	var sub string
	if len(subName) > 0 {
		sub = subName[0]
	}
	if _, err = session.Run("DELETE FROM ejaSession WHERE ejaOwner=? AND name=? AND sub=?", userId, name, sub); err != nil {
		return
	}
	if _, err = session.Run("INSERT INTO ejaSession (ejaId, ejaOwner, ejaLog, name, value, sub) VALUES (NULL,?,?,?,?,?)", userId, session.Now(), name, value, sub); err != nil {
		return
	}
	if _, err = session.Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name=? AND sub=?", userId, name, sub); err != nil {
		return
	}
	if _, err = session.Run("INSERT INTO ejaSessions (ejaId, ejaOwner, ejaLog, name, value, sub) VALUES (NULL,?,?,?,?,?)", userId, session.Now(), name, value, sub); err != nil {
		return err
	}
	return
}

// SessionCleanLink removes specific session variables related to links for a specified user.
func (session *TypeSession) SessionCleanLink(userId int64) error {
	_, err := session.Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name in ('Link','SqlQuery64','SqlQueryArgs','SearchLimit','SearchOffset','SearchOrder')", userId)
	return err
}

// SessionCleanSearch removes specific session variables related to searches for a specified user.
func (session *TypeSession) SessionCleanSearch(userId int64) error {
	_, err := session.Run("DELETE FROM ejaSessions WHERE ejaOwner=? AND name in ('SqlQuery64','SqlQueryArgs','SearchLimit','SearchOffset','SearchOrder')", userId)
	return err
}

// SessionReset removes all session variables for a specified user from the database.
func (session *TypeSession) SessionReset(userId int64) error {
	if _, err := session.Run("DELETE FROM ejaSessions WHERE ejaOwner=?", userId); err != nil {
		return err
	}
	if _, err := session.Run("UPDATE ejaUsers SET ejaSession='' WHERE ejaId=?", userId); err != nil {
		return err
	}
	return nil
}
