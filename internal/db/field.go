// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"errors"
	"fmt"
)

// TypeField represents a field in a database with metadata.
type TypeField struct {
	Type    string
	Name    string
	Label   string
	Value   string
	Options []TypeSelect
}

// FieldNameList retrieves a list of field names based on the provided module ID and action type.
func FieldNameList(moduleId int64, actionType string) (fields []string) {
	rows, err := Rows(fmt.Sprintf("SELECT name FROM ejaFields WHERE ejaModuleId=%d AND power%s>0 AND power%s!='' ORDER BY power%s ASC;", moduleId, actionType, actionType, actionType))
	if err == nil {
		for _, row := range rows {
			fields = append(fields, row["name"])
		}
	}
	return
}

// Fields retrieves a list of TypeField objects based on the provided module ID, action type, and field values.
func Fields(ownerId int64, moduleId int64, actionType string, values map[string]string) ([]TypeField, error) {
	var res []TypeField

	rows, err := Rows(fmt.Sprintf("SELECT * FROM ejaFields WHERE ejaModuleId=%d AND power%s>0 AND power%s!='' ORDER BY power%s ASC;", moduleId, actionType, actionType, actionType))
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
			rowOptions = SelectToRows(row["value"])
		}
		if rowType == "sqlMatrix" {
			rowOptions = SelectSqlToRows(row["value"])
		}
		if rowType == "sqlValue" || rowType == "sqlHidden" {
			rowValue, _ = Value(row["value"])
		}

		if Number(row["translate"]) > 0 {
			rowValue = Translate(rowValue, ownerId)
		}

		if actionType == "Edit" && rowName == "ejaOwner" && Number(rowValue) < 1 {
			rowValue = String(ownerId)
		}

		field := TypeField{
			Type:    rowType,
			Name:    rowName,
			Label:   Translate(rowName, ownerId),
			Value:   rowValue,
			Options: rowOptions,
		}
		res = append(res, field)
	}

	return res, nil
}

// FieldAdd adds a new field to the specified table with the given name and type.
func FieldAdd(tableName string, fieldName string, fieldType string) error {
	check, err := TableExists(tableName)
	if err != nil {
		return err
	}
	if !check {
		return errors.New("table does not exist")
	}

	check, err = FieldExists(tableName, fieldName)
	if err != nil {
		return err
	}
	if check {
		return errors.New("field already exists")
	}

	sqlFieldType := FieldType(fieldType)
	_, err = Run(fmt.Sprintf("ALTER TABLE %s ADD %s %s", tableName, fieldName, sqlFieldType))
	if err != nil {
		return err
	}

	return nil
}

// FieldExists checks whether a field with the given name already exists in the specified table.
func FieldExists(tableName string, fieldName string) (bool, error) {
	switch DbEngine {
	case "sqlite":
		return sqliteFieldExists(tableName, fieldName)
	case "mysql":
		return mysqlFieldExists(tableName, fieldName)
	default:
		return false, errors.New("engine not found")
	}
}

// FieldNameIsValid checks the validity of a field name based on the current database engine.
func FieldNameIsValid(name string) error {
	switch DbEngine {
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
func FieldTypeGet(moduleId int64, fieldName string) string {
	value, _ := Value("SELECT type FROM ejaFields WHERE ejaModuleId=? AND name=?", moduleId, fieldName)
	return value
}
