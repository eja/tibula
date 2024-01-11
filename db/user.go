// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// UserGetAllByUserAndPass retrieves user information based on the provided username and hashed password.
func UserGetAllByUserAndPass(username string, password string) TypeRow {
	result, _ := Row("SELECT * FROM ejaUsers WHERE username=? AND password=?", username, Password(password))
	return result
}

// UserGetAllById retrieves all user information based on the provided user ID.
func UserGetAllById(userId int64) TypeRow {
	result, _ := Row("SELECT * FROM ejaUsers WHERE ejaId=?", userId)
	return result
}

// UserGetAllBySession retrieves user information based on the provided session.
func UserGetAllBySession(session string) TypeRow {
	result, _ := Row("SELECT * FROM ejaUsers WHERE ejaSession=?", session)
	return result
}

// UserPermissionCopy copies user permissions from one module to another.
func UserPermissionCopy(userId int64, moduleId int64) {
	Run(`
		INSERT INTO ejaLinks (ejaId, ejaOwner, ejaLog, srcModuleId, srcFieldId, dstModuleId, dstFieldId, power)
		SELECT NULL, 1, ?, ?, ejaId, ?, ?, 2
		FROM ejaPermissions
		WHERE ejaModuleId = ?;
		`, Now(), ModuleGetIdByName("ejaPermissions"), ModuleGetIdByName("ejaUsers"), userId, moduleId)
}
