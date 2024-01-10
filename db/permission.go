// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// PermissionCount returns the count of permissions associated with the specified module.
func PermissionCount(moduleId int64) int64 {
	value, _ := Value("SELECT COUNT(*) FROM ejaPermissions WHERE ejaModuleId=?", moduleId)
	return Number(value)
}

// PermissionAddDefault adds default permissions for the specified user and module.
func PermissionAddDefault(userId int64, moduleId int64) int64 {
	check, _ := Run(`
		INSERT INTO ejaPermissions 
			(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaCommandId) 
		SELECT 
			NULL,?,?,?,ejaId FROM ejaCommands WHERE defaultCommand>0
		`, userId, Now(), moduleId)
	return check.Changes
}

// PermissionAdd adds a permission for the specified user, module, and command.
func PermissionAdd(userId int64, moduleId int64, commandName string) int64 {
	check, _ := Run(`
		INSERT INTO ejaPermissions 
			(ejaId, ejaOwner, ejaLog, ejaModuleId, ejaCommandId) 
		SELECT NULL,?,?,?,ejaId FROM ejaCommands WHERE name=?
		`, userId, Now(), moduleId, commandName)
	return check.Changes
}
