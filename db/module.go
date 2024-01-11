// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// TypeModule represents a modular structure containing information about modules, fields, translations, and data.
type TypeModule struct {
	Module      TypeModuleModule         `json:"module"`
	Command     []string                 `json:"command"`
	Field       []TypeModuleField        `json:"field"`
	Translation []TypeModuleTranslation  `json:"translation"`
	Name        string                   `json:"name"`
	Data        []map[string]interface{} `json:"data"`
}

// TypeModuleModule represents metadata about a module within a TypeModule.
type TypeModuleModule struct {
	ParentName  string `json:"parentName,omitempty"`
	Power       int    `json:"power"`
	SearchLimit int    `json:"searchLimit"`
	SqlCreated  int    `json:"sqlCreated"`
	SortList    string `json:"sortList,omitempty"`
}

// TypeModuleField represents metadata about a field within a TypeModule.
type TypeModuleField struct {
	Value       string `json:"value"`
	PowerEdit   int    `json:"powerEdit"`
	PowerList   int    `json:"powerList"`
	Type        string `json:"type"`
	Translate   int    `json:"translate"`
	PowerSearch int    `json:"powerSearch"`
	Name        string `json:"name"`
}

// TypeModuleTranslation represents translation information within a TypeModule.
type TypeModuleTranslation struct {
	EjaLanguage   string `json:"ejaLanguage"`
	EjaModuleName string `json:"ejaModuleName,omitempty"`
	Word          string `json:"word"`
	Translation   string `json:"translation"`
}

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
