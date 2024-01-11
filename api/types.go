// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"github.com/eja/tibula/db"
)

type TypeApi struct {
	Action             string
	ActionType         string
	Alert              []string
	Commands           []db.TypeCommand
	DefaultSearchLimit int64
	DefaultSearchOrder string
	Export             db.TypeModule
	FieldNameList      []string
	Fields             []db.TypeField
	Id                 int64
	IdList             []int64
	Info               []string
	Language           string
	Link               db.TypeLink
	Linking            bool
	Links              []db.TypeLink
	ModuleId           int64
	ModuleLabel        string
	ModuleName         string
	Owner              int64
	Path               []db.TypeModulePath
	SearchCols         []string
	SearchCount        int64
	SearchLabels       map[string]string
	SearchLast         int64
	SearchLimit        int64
	SearchLink         bool
	SearchLinkClean    bool
	SearchLinks        []string
	SearchOffset       int64
	SearchOrder        map[string]string
	SearchRows         db.TypeRows
	Session            string
	SqlQuery           string
	SqlQuery64         string
	SqlQueryArgs       []interface{}
	Tree               []db.TypeModuleTree
	Values             map[string]string
}
