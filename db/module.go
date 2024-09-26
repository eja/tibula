// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

// TypeModule represents a modular structure containing information about modules, fields, translations, links and data.
type TypeModule struct {
	Module      TypeModuleModule         `json:"module"`
	Command     []string                 `json:"command"`
	Field       []TypeModuleField        `json:"field"`
	Link        []TypeModuleLink         `json:"link,omitempty"`
	Translation []TypeModuleTranslation  `json:"translation,omitempty"`
	Name        string                   `json:"name"`
	Data        []map[string]interface{} `json:"data,omitempty"`
	Type        string                   `json:"type"`
}

// TypeModuleModule represents metadata about a module within a TypeModule.
type TypeModuleModule struct {
	ParentName  string `json:"parentName,omitempty"`
	Power       int64  `json:"power"`
	SearchLimit int64  `json:"searchLimit"`
	SqlCreated  int64  `json:"sqlCreated"`
	SortList    string `json:"sortList,omitempty"`
}

// TypeModuleField represents metadata about a field within a TypeModule.
type TypeModuleField struct {
	Value       string `json:"value"`
	PowerEdit   int64  `json:"powerEdit"`
	PowerList   int64  `json:"powerList"`
	Type        string `json:"type"`
	Translate   int64  `json:"translate"`
	PowerSearch int64  `json:"powerSearch"`
	Name        string `json:"name"`
	SizeSearch  int64  `json:"sizeSearch"`
	SizeList    int64  `json:"sizeList"`
	SizeEdit    int64  `json:"sizeEdit"`
}

// TypeModuleTranslation represents translation information within a TypeModule.
type TypeModuleTranslation struct {
	EjaLanguage   string `json:"ejaLanguage"`
	EjaModuleName string `json:"ejaModuleName,omitempty"`
	Word          string `json:"word"`
	Translation   string `json:"translation"`
}

// TypeModuleModule represents module links withing modules in TypeModule.
type TypeModuleLink struct {
	SrcField  string `json:"srcField,omitempty"`
	SrcModule string `json:"srcModule"`
	DstModule string `json:"dstModule"`
	Power     int64  `json:"power,omitempty"`
}

// ModuleGetIdByName retrieves the module ID based on the given module name.
// If an error occurs during the database operation or table name is not valid, it returns 0.
func (session *TypeSession) ModuleGetIdByName(name string) int64 {
	if err := session.TableNameIsValid(name); err != nil {
		return 0
	}

	val, err := session.Value("SELECT ejaId FROM ejaModules WHERE name=?", name)
	if err != nil {
		return 0
	}
	return session.Number(val)
}

// ModuleGetNameById retrieves the module name based on the given module ID.
// If an error occurs during the database operation or table name is not valid, it returns an empty string.
func (session *TypeSession) ModuleGetNameById(id int64) string {
	val, err := session.Value("SELECT name FROM ejaModules WHERE ejaId = ?", id)
	if err != nil {
		return ""
	}
	name := session.String(val)
	if err := session.TableNameIsValid(name); err != nil {
		return ""
	}
	return name
}
