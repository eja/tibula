// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// PermissionCount returns the count of permissions associated with the specified module.
func (session *TypeSession) PermissionCount(moduleId int64) int64 {
	value, _ := session.Value("SELECT COUNT(*) FROM ejaPermissions WHERE ejaModuleId=?", moduleId)
	return session.Number(value)
}

// PermissionAddDefault adds default permissions for the specified user and module.
func (session *TypeSession) PermissionAddDefault(userId int64, moduleId int64) int64 {
	check, _ := session.Run(`
		INSERT INTO ejaPermissions 
			(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaCommandId) 
		SELECT 
			NULL,?,?,?,ejaId FROM ejaCommands WHERE defaultCommand>0
		`, userId, session.Now(), moduleId)
	return check.Changes
}

// PermissionAdd adds a permission for the specified user, module, and command.
func (session *TypeSession) PermissionAdd(userId int64, moduleId int64, commandName string) int64 {
	check, _ := session.Run(`
		INSERT INTO ejaPermissions 
			(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaCommandId) 
		SELECT NULL,?,?,?,ejaId FROM ejaCommands WHERE name=?
		`, userId, session.Now(), moduleId, commandName)
	return check.Changes
}
