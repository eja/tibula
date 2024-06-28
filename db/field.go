// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

// TypeField represents a field in a database with metadata.
type TypeField struct {
	Type        string
	Name        string
	Label       string
	Value       string
	SearchIndex int64
	SearchSize  int64
	ListIndex   int64
	ListSize    int64
	EditIndex   int64
	EditSize    int64
	Options     []TypeSelect
}

// FieldNameList retrieves a list of field names based on the provided module ID and action type.
func (session *TypeSession) FieldNameList(moduleId int64, actionType string) (fields []string) {
	rows, err := session.Rows(fmt.Sprintf("SELECT name FROM ejaFields WHERE ejaModuleId=%d AND power%s>0 AND power%s!='' ORDER BY power%s ASC;", moduleId, actionType, actionType, actionType))
	if err == nil {
		for _, row := range rows {
			fields = append(fields, row["name"])
		}
	}
	return
}

// Fields retrieves a list of TypeField objects based on the provided module ID, action type, and field values.
func (session *TypeSession) Fields(ownerId int64, moduleId int64, actionType string, values map[string]string) ([]TypeField, error) {
	var res []TypeField

	rows, err := session.Rows(fmt.Sprintf("SELECT * FROM ejaFields WHERE ejaModuleId=%d AND power%s>0 AND power%s!='' ORDER BY power%s ASC;", moduleId, actionType, actionType, actionType))
	if err != nil {
		return res, err
	}

	for _, row := range rows {
		rowType := row["type"]
		rowName := row["name"]
		var rowValue string
		var rowOptions []TypeSelect

		if values[rowName] != "" {
			rowValue = values[rowName]
		} else if row["value"] != "" {
			rowValue = row["value"]
		}

		if rowType == "select" {
			rowOptions = session.SelectToRows(row["value"])
		}
		if rowType == "sqlMatrix" {
			rowOptions = session.SelectSqlToRows(row["value"])
		}
		if rowType == "sqlValue" || rowType == "sqlHidden" {
			rowValue, _ = session.Value(row["value"])
		}

		if session.Number(row["translate"]) > 0 {
			rowValue = session.Translate(rowValue, ownerId)
		}

		if actionType == "Edit" && rowName == "ejaOwner" && session.Number(rowValue) < 1 {
			rowValue = session.String(ownerId)
		}

		field := TypeField{
			Type:        rowType,
			Name:        rowName,
			Label:       session.Translate(rowName, ownerId),
			Value:       rowValue,
			Options:     rowOptions,
			SearchIndex: session.Number(row["powerSearch"]),
			SearchSize:  session.Number(row["sizeSearch"]),
			ListIndex:   session.Number(row["powerList"]),
			ListSize:    session.Number(row["sizeList"]),
			EditIndex:   session.Number(row["powerEdit"]),
			EditSize:    session.Number(row["sizeEdit"]),
		}
		res = append(res, field)
	}

	return res, nil
}

// FieldAdd adds a new field to the specified table with the given name and type.
func (session *TypeSession) FieldAdd(tableName string, fieldName string, fieldType string) error {
	if fieldType == "label" || fieldType == "sqlValue" {
		return nil
	}

	check, err := session.TableExists(tableName)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table does not exist")
	}

	check, err = session.FieldExists(tableName, fieldName)
	if err != nil {
		return err
	}
	if check {
		return errors.New("field already exists")
	}

	sqlFieldType := FieldType(fieldType)
	_, err = session.Run(fmt.Sprintf("ALTER TABLE %s ADD %s %s", tableName, fieldName, sqlFieldType))
	if err != nil {
		return err
	}

	return nil
}

// FieldExists checks whether a field with the given name already exists in the specified table.
func (session *TypeSession) FieldExists(tableName string, fieldName string) (bool, error) {
	switch session.Engine {
	case "sqlite":
		return session.sqliteFieldExists(tableName, fieldName)
	case "mysql":
		return session.mysqlFieldExists(tableName, fieldName)
	default:
		return false, errors.New("engine not found")
	}
}

// FieldNameIsValid checks the validity of a field name based on the current database engine.
func (session *TypeSession) FieldNameIsValid(name string) error {
	switch session.Engine {
	case "sqlite":
		return sqliteFieldNameIsValid(name)
	case "mysql":
		return mysqlFieldNameIsValid(name)
	default:
		return errors.New("engine not found")
	}
}

// FieldType returns the corresponding SQL type for a given field type.
func FieldType(name string) string {
	switch name {
	case "boolean", "integer":
		return "INTEGER"
	case "decimal":
		return "DOUBLE"
	case "date":
		return "DATE"
	case "time":
		return "TIME"
	case "datetime":
		return "DATETIME"
	default:
		return "TEXT"
	}
}

// FieldTypeGet retrieves the field type for a specific field in a module based on module ID and field name.
func (session *TypeSession) FieldTypeGet(moduleId int64, fieldName string) string {
	value, _ := session.Value("SELECT type FROM ejaFields WHERE ejaModuleId=? AND name=?", moduleId, fieldName)
	return value
}
