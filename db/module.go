// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// ModuleGetIdByName retrieves the module ID based on the given module name.
// If an error occurs during the database operation or table name is not valid, it returns 0.
func ModuleGetIdByName(name string) int64 {
	if err := TableNameIsValid(name); err != nil {
		return 0
	}

	val, err := Value("SELECT ejaId FROM ejaModules WHERE name=?", name)
	if err != nil {
		return 0
	}
	return Number(val)
}

// ModuleGetNameById retrieves the module name based on the given module ID.
// If an error occurs during the database operation or table name is not valid, it returns an empty string.
func ModuleGetNameById(id int64) string {
	val, err := Value("SELECT name FROM ejaModules WHERE ejaId = ?", id)
	if err != nil {
		return ""
	}
	name := String(val)
	if err := TableNameIsValid(name); err != nil {
		return ""
	}
	return name
}
