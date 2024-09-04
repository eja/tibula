// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// UserGetAllByUserAndPass retrieves user information based on the provided username and hashed password.
func (session *TypeSession) UserGetAllByUserAndPass(username string, password string) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE username=? AND password=?", username, session.Password(password))
	return result
}

// UserGetAllById retrieves all user information based on the provided user ID.
func (session *TypeSession) UserGetAllById(userId int64) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE ejaId=?", userId)
	return result
}

// UserGetAllBySession retrieves user information based on the provided session.
func (session *TypeSession) UserGetAllBySession(sessionHash string) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE ejaSession=?", sessionHash)
	return result
}

// UserGetAllByUsername retrieves user information based on the provided username.
func (session *TypeSession) UserGetAllByUsername(username string) TypeRow {
	result, _ := session.Row("SELECT * FROM ejaUsers WHERE username=?", username)
	return result
}

// UserPermissionCopy copies user permissions from one module to another.
func (session *TypeSession) UserPermissionCopy(userId int64, moduleId int64) {
	session.Run(`
		INSERT INTO ejaLinks (ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power)
		SELECT NULL, 1, ?, ?, ejaId, ?, ?, 2
		FROM ejaPermissions
		WHERE ejaModuleId = ?;
		`, session.Now(), session.ModuleGetIdByName("ejaPermissions"), session.ModuleGetIdByName("ejaUsers"), userId, moduleId)
}
