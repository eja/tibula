// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
	"regexp"
	"strings"
)

// TypeSearchColumn represents a mapping of column names to their properties and values.
type TypeSearchColumn map[string]map[string]interface{}

// SearchMatrix performs a search on a specified database module using a dynamic SQL query.
func SearchMatrix(ownerId int64, moduleId int64, query string, queryArgs []interface{}) (resultRows TypeRows, resultCols []string, resultLabels map[string]string, err error) {
	var sqlResult TypeRows
	var queryHead TypeSearchColumn
	resultLabels = make(map[string]string)

	queryHead, query, err = searchHeader(query, moduleId)
	if err != nil {
		return
	}

	sqlResult, err = Rows(query, queryArgs...)
	if err != nil {
		return
	}

	resultCols, err = Cols(query, queryArgs...)
	if err != nil {
		return
	}

	for _, val := range resultCols {
		resultLabels[val] = Translate(val, ownerId)
	}

	for _, row := range sqlResult {
		resultRows = append(resultRows, searchRow(ownerId, queryHead, row))
	}

	return
}

// SearchQuery generates a SQL query for searching records in a specified table based on provided criteria.
func SearchQuery(ownerId int64, tableName string, values map[string]string) (string, []interface{}, error) {
	var sql []string
	var args []interface{}

	moduleId := ModuleGetIdByName(tableName)
	sqlType := make(map[string]string)

	sql = append(sql, "SELECT ejaId")

	rows, err := Rows("SELECT * FROM ejaFields WHERE ejaModuleId=? ORDER BY powerList", moduleId)
	if err != nil {
		return "", nil, err
	}
	for _, row := range rows {
		if Number(row["powerList"]) > 0 {
			sql = append(sql, ",")
			sql = append(sql, row["name"])
		}
		sqlType[row["name"]] = row["type"]
	}

	sql = append(sql, fmt.Sprintf(" FROM %s WHERE ejaOwner IN ("+NumbersToCsv(Owners(ownerId, moduleId))+") ", tableName))

	for keyRaw, val := range values {
		keyMode := ""
		keySplit := strings.Split(keyRaw, ".")
		key := keySplit[0]
		if len(keySplit) == 2 {
			keyMode = keySplit[1]
		}
		if FieldNameIsValid(key) == nil && val != "" {
			sqlTypeThis := String(sqlType[key])
			sqlAnd := ""
			arg := ""

			switch sqlTypeThis {
			case "boolean":
				sqlAnd = fmt.Sprintf(" AND %s = ? ", key)
			case "integer", "decimal":
				switch keyMode {
				case "start":
					sqlAnd = fmt.Sprintf(" AND %s >= ? ", key)
				case "stop":
					sqlAnd = fmt.Sprintf(" AND %s <= ? ", key)
				default:
					sqlAnd = fmt.Sprintf(" AND %s = ? ", key)
				}
			case "date", "time", "datetime":
				switch keyMode {
				case "start":
					sqlAnd = fmt.Sprintf(" AND %s >= ? ", key)
				case "stop":
					sqlAnd = fmt.Sprintf(" AND %s <= ? ", key)
				default:
					sqlAnd = fmt.Sprintf(" AND %s = ? ", key)
				}
			}
			if sqlAnd == "" {
				sqlAnd = fmt.Sprintf(" AND %s LIKE ? ", key)
			}
			if arg == "" {
				arg = val
			}

			if sqlAnd != "" && arg != "" {
				sql = append(sql, sqlAnd)
				args = append(args, arg)
			}
		}
	}
	return strings.Join(sql, ""), args, nil
}

// SearchCount calculates the number of records for a given search query and arguments.
func SearchCount(query string, args []interface{}) int64 {
	var queryCount string
	start := strings.Index(strings.ToUpper(query), "FROM")
	stop := strings.LastIndex(strings.ToUpper(query), "LIMIT")
	if start > 0 {
		if stop > 0 {
			queryCount = query[start:stop]
		} else {
			queryCount = query[start:]
		}
		if result, err := Value("SELECT COUNT(*) "+queryCount, args...); err != nil {
			return 0
		} else {
			return Number(result)
		}
	}
	return 0
}

// searchHeader retrieves column information for constructing search queries.
func searchHeader(query string, moduleId int64) (TypeSearchColumn, string, error) {
	colValues := make(TypeSearchColumn)

	rows, err := Rows("SELECT * FROM ejaFields WHERE ejaModuleId=? AND powerList>0 ORDER BY powerList", moduleId)
	if err != nil {
		return nil, "", err
	}

	for _, field := range rows {
		rowType := field["type"]
		rowName := field["name"]
		colValues[rowName] = make(map[string]interface{})
		colValues[rowName]["type"] = rowType
		switch rowType {
		case "boolean":
			colValues[rowName]["value"] = []TypeSelect{{Key: "0", Value: "FALSE"}, {Key: "1", Value: "TRUE"}}
		case "select":
			colValues[rowName]["value"] = SelectToRows(field["value"])
		case "sqlMatrix":
			colValues[rowName]["value"] = SelectSqlToRows(field["value"])
		case "sqlValue", "sqlHidden":
			query = strings.Replace(query, field["name"], fmt.Sprintf("(%s) AS %s", field["value"], field["name"]), 1)
		}

		colValues[field["name"]]["translation"] = 0
		if Number(field["translate"]) > 0 {
			colValues[field["name"]]["translation"] = 1
		}
	}
	return colValues, query, nil
}

// searchRow filters and transforms a row based on the search column information.
func searchRow(ownerId int64, queryHead TypeSearchColumn, row TypeRow) TypeRow {
	filteredRow := row
	for colName := range queryHead {
		if queryHead[colName]["value"] != nil && row[colName] != "" {
			value := row[colName]
			if strings.HasPrefix(colName, "ejaId") {
				value = row["ejaId"]
			} else {
				if colValues, ok := queryHead[colName]["value"].([]TypeSelect); ok {
					for _, colValue := range colValues {
						if colValue.Key == value {
							value = colValue.Value
						}
					}
				}
			}
			filteredRow[colName] = value
		}
		if len(row[colName]) >= 19 {
			switch queryHead[colName]["type"] {
			case "datetime":
				filteredRow[colName] = row[colName][:10] + " " + row[colName][11:19]
			case "date":
				filteredRow[colName] = row[colName][:10]
			case "time":
				filteredRow[colName] = row[colName][11:19]
			}
		}
		if Number(queryHead[colName]["translation"]) > 0 {
			filteredRow[colName] = Translate(row[colName], ownerId)
		}
	}

	return filteredRow
}

// SearchQueryOrderAndLimit generates an ORDER BY, LIMIT, and OFFSET clause for search queries.
func SearchQueryOrderAndLimit(order string, limit int64, offset int64) string {
	pattern := `^\s*(\w+\s+(ASC|DESC)\s*,\s*)*\w+\s+(ASC|DESC)\s*$`
	regexpPattern := regexp.MustCompile(pattern)
	if !regexpPattern.MatchString(order) {
		Warn("order by is not regex compatible", order)
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("ORDER BY %s LIMIT %d OFFSET %d", order, limit, offset)
}

// SearchQueryLinks generates a condition for searching based on related links.
func SearchQueryLinks(ownerId, srcModuleId, srcFieldId, dstModuleId int64) string {
	result := ""
	links := SearchLinks(ownerId, srcModuleId, srcFieldId, dstModuleId)
	if len(links) > 0 {
		result = " AND ejaId IN (" + strings.Join(links, ",") + ") "
	}
	return result
}
