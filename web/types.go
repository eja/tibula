// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"github.com/eja/tibula/db"
)

type TypeEja struct {
	Action             string
	ActionType         string
	Commands           []db.TypeCommand
	DefaultSearchOrder string
	DefaultSearchLimit int64
	Fields             []db.TypeField
	FieldNameList      []string
	Id                 int64
	IdList             []int64
	Info               []string
	Alert              []string
	Language           string
	ModuleId           int64
	ModuleName         string
	ModuleLabel        string
	Owner              int64
	SearchLimit        int64
	SearchOffset       int64
	SearchLast         int64
	SearchCount        int64
	SearchRows         db.TypeRows
	SearchCols         []string
	SearchOrder        map[string]string
	SearchLabels       map[string]string
	SearchLinks        []string
	SearchLink         bool
	SearchLinkClean    bool
	Session            string
	SqlQuery           string
	SqlQuery64         string
	SqlQueryArgs       []interface{}
	Values             map[string]string
	Path               []db.TypeModulePath
	Tree               []db.TypeModuleTree
	Link               db.TypeLink
	Links              []db.TypeLink
	Linking            bool
}
