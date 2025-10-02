// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

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

type TypeModuleModule struct {
	ParentName  string `json:"parentName,omitempty"`
	Power       int64  `json:"power"`
	SearchLimit int64  `json:"searchLimit"`
	SqlCreated  int64  `json:"sqlCreated"`
	SortList    string `json:"sortList,omitempty"`
}

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

type TypeModuleTranslation struct {
	EjaLanguage   string `json:"ejaLanguage"`
	EjaModuleName string `json:"ejaModuleName,omitempty"`
	Word          string `json:"word"`
	Translation   string `json:"translation"`
}

type TypeModuleLink struct {
	SrcField  string `json:"srcField,omitempty"`
	SrcModule string `json:"srcModule"`
	DstModule string `json:"dstModule"`
	Power     int64  `json:"power,omitempty"`
}

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
