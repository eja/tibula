// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"github.com/eja/tibula/db"
)

type TypeApi struct {
	Action              string              `json:"Action,omitempty"`
	ActionType          string              `json:"ActionType,omitempty"`
	Alert               []string            `json:"Alert,omitempty"`
	Commands            []db.TypeCommand    `json:"Commands,omitempty"`
	DefaultSearchLimit  int64               `json:"DefaultSearchLimit,omitempty"`
	DefaultSearchOrder  string              `json:"DefaultSearchOrder,omitempty"`
	FieldNameList       []string            `json:"FieldNameList,omitempty"`
	Fields              []db.TypeField      `json:"Fields,omitempty"`
	Id                  int64               `json:"Id,omitempty"`
	IdList              []int64             `json:"IdList,omitempty"`
	Info                []string            `json:"Info,omitempty"`
	Language            string              `json:"Language,omitempty"`
	Link                db.TypeLink         `json:"Link,omitempty"`
	Linking             bool                `json:"Linking,omitempty"`
	Links               []db.TypeLink       `json:"Links,omitempty"`
	ModuleId            int64               `json:"ModuleId,omitempty"`
	ModuleLabel         string              `json:"ModuleLabel,omitempty"`
	ModuleName          string              `json:"ModuleName,omitempty"`
	Owner               int64               `json:"Owner,omitempty"`
	Path                []db.TypeModulePath `json:"Path,omitempty"`
	SearchCols          []string            `json:"SearchCols,omitempty"`
	SearchCount         int64               `json:"SearchCount,omitempty"`
	SearchLabels        map[string]string   `json:"SearchLabels,omitempty"`
	SearchLast          int64               `json:"SearchLast,omitempty"`
	SearchLimit         int64               `json:"SearchLimit,omitempty"`
	SearchLink          bool                `json:"SearchLink,omitempty"`
	SearchLinkClean     bool                `json:"SearchLinkClean,omitempty"`
	SearchLinks         []string            `json:"SearchLinks,omitempty"`
	SearchOffset        int64               `json:"SearchOffset,omitempty"`
	SearchOrder         map[string]string   `json:"SearchOrder,omitempty"`
	SearchRows          db.TypeRows         `json:"SearchRows,omitempty"`
	Session             string              `json:"Session,omitempty"`
	SqlQuery            string              `json:"SqlQuery,omitempty"`
	SqlQuery64          string              `json:"SqlQuery64,omitempty"`
	SqlQueryArgs        []interface{}       `json:"SqlQueryArgs,omitempty"`
	Tree                []db.TypeModuleTree `json:"Tree,omitempty"`
	Values              map[string]string   `json:"Values,omitempty"`
	GoogleSsoId         string              `json:"GoogleSsoId,omitempty"`
	SubModules          []db.TypeLink       `json:"SubModules,omitempty"`
	SubModulePath       []SubModulePathItem `json:"SubModulePath,omitempty"`
	SubModulePathString string              `json:"SubModulePathString,omitempty"`
}

type SubModulePathItem struct {
	LinkingModuleId int64
	ModuleId        int64
	FieldId         int64
	FieldName       string
}
