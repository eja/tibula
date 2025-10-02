// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
	"time"
)

func (session *TypeSession) UserGetAllByUserAndPass(username string, password string) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE username=? AND password=?", username, session.Password(password))
	return result
}

func (session *TypeSession) UserGetAllById(userId int64) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE ejaId=?", userId)
	return result
}

func (session *TypeSession) UserGetAllBySession(sessionHash string) (result TypeRow) {
	timeNow := time.Now().Unix() / 1000
	timePre := timeNow - 1
	rows, err := session.Rows("SELECT * FROM ejaUsers")
	if err == nil {
		for _, row := range rows {
			hashNow := session.Sha256(fmt.Sprintf("%s.%s.%d", row["ejaSession"], row["ejaId"], timeNow))
			hashPre := session.Sha256(fmt.Sprintf("%s.%s.%d", row["ejaSession"], row["ejaId"], timePre))
			if hashNow == sessionHash || hashPre == sessionHash {
				return row
			}
		}
	}
	return
}

func (session *TypeSession) UserGetAllByUsername(username string) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE username=?", username)
	return result
}

func (session *TypeSession) UserPermissionCopy(userId int64, moduleId int64) {
	session.Run(`
		INSERT INTO ejaLinks (ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power)
		SELECT NULL, 1, ?, ?, ejaId, ?, ?, 2
		FROM ejaPermissions
		WHERE ejaModuleId = ?;
		`, session.Now(), session.ModuleGetIdByName("ejaPermissions"), session.ModuleGetIdByName("ejaUsers"), userId, moduleId)
}
