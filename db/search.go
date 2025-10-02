// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eja/tibula/log"
)

type TypeSearchColumn map[string]map[string]interface{}

func (session *TypeSession) SearchMatrix(ownerId int64, moduleId int64, query string, queryArgs []interface{}) (resultRows TypeRows, resultCols []string, resultLabels map[string]string, err error) {
	var sqlResult TypeRows
	var queryHead TypeSearchColumn
	resultLabels = make(map[string]string)

	queryHead, query, err = session.searchHeader(query, moduleId)
	if err != nil {
		return
	}

	sqlResult, err = session.Rows(query, queryArgs...)
	if err != nil {
		return
	}

	resultCols, err = session.Cols(query, queryArgs...)
	if err != nil {
		return
	}

	for _, val := range resultCols {
		resultLabels[val] = session.Translate(val, ownerId)
	}

	for _, row := range sqlResult {
		resultRows = append(resultRows, session.searchRow(ownerId, queryHead, row))
	}

	return
}

func (session *TypeSession) SearchQuery(ownerId int64, tableName string, values map[string]string) (string, []interface{}, error) {
	var sql []string
	var args []interface{}

	moduleId := session.ModuleGetIdByName(tableName)
	sqlType := make(map[string]string)

	sql = append(sql, "SELECT ejaId")

	rows, err := session.Rows("SELECT * FROM ejaFields WHERE ejaModuleId=? AND type NOT IN ('label') ORDER BY powerList", moduleId)
	if err != nil {
		return "", nil, err
	}
	for _, row := range rows {
		if session.Number(row["powerList"]) > 0 {
			sql = append(sql, ",")
			sql = append(sql, row["name"])
		}
		sqlType[row["name"]] = row["type"]
	}

	sql = append(sql, fmt.Sprintf(" FROM %s WHERE ejaOwner IN ("+session.NumbersToCsv(session.Owners(ownerId, moduleId))+") ", tableName))

	for keyRaw, val := range values {
		keyMode := ""
		keySplit := strings.Split(keyRaw, ".")
		key := keySplit[0]
		if len(keySplit) == 2 {
			keyMode = keySplit[1]
		}
		if session.FieldNameIsValid(key) == nil && val != "" {
			sqlTypeThis := session.String(sqlType[key])
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
			case "multiple", "sqlMultiple":
				sqlAnd = fmt.Sprintf(` AND %s LIKE '%%"' || ? || '"%%' `, key)
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

func (session *TypeSession) SearchCount(query string, args []interface{}) int64 {
	reFrom := regexp.MustCompile(`(?i)\s+FROM\s+`)
	reLimit := regexp.MustCompile(`(?i)\s+LIMIT\s+`)

	fromMatches := reFrom.FindAllStringIndex(query, -1)
	if len(fromMatches) == 0 {
		return 0
	}

	fromPos := fromMatches[len(fromMatches)-1][1] - 5

	limitMatches := reLimit.FindAllStringIndex(query, -1)
	var queryCount string

	if len(limitMatches) > 0 {
		queryCount = query[fromPos:limitMatches[0][0]]
	} else {
		queryCount = query[fromPos:]
	}

	if result, err := session.Value("SELECT COUNT(*) "+queryCount, args...); err != nil {
		return 0
	} else {
		return session.Number(result)
	}
}

func (session *TypeSession) searchHeader(query string, moduleId int64) (TypeSearchColumn, string, error) {
	colValues := make(TypeSearchColumn)

	rows, err := session.Rows("SELECT * FROM ejaFields WHERE ejaModuleId=? AND powerList>0 ORDER BY powerList", moduleId)
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
		case "select", "multiple":
			colValues[rowName]["value"] = session.SelectToRows(field["value"])
		case "sqlMatrix", "sqlMultiple":
			colValues[rowName]["value"] = session.SelectSqlToRows(field["value"])
		case "sqlValue", "sqlHidden":
			query = strings.Replace(query, field["name"], fmt.Sprintf("(%s) AS %s", field["value"], field["name"]), 1)
		}

		colValues[field["name"]]["translation"] = 0
		if session.Number(field["translate"]) > 0 {
			colValues[field["name"]]["translation"] = 1
		}
	}
	return colValues, query, nil
}

func (session *TypeSession) searchRow(ownerId int64, queryHead TypeSearchColumn, row TypeRow) TypeRow {
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
		if session.Number(queryHead[colName]["translation"]) > 0 {
			filteredRow[colName] = session.Translate(row[colName], ownerId)
		}
	}

	return filteredRow
}

func (session *TypeSession) SearchQueryOrderAndLimit(order string, limit int64, offset int64) string {
	pattern := `^\s*(\w+\s+(ASC|DESC)\s*,\s*)*\w+\s+(ASC|DESC)\s*$`
	regexpPattern := regexp.MustCompile(pattern)
	if !regexpPattern.MatchString(order) {
		log.Warn(tag, "order by is not regex compatible", order)
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("ORDER BY %s LIMIT %d OFFSET %d", order, limit, offset)
}

func (session *TypeSession) SearchQueryLinks(ownerId, srcModuleId, srcFieldId, dstModuleId int64) string {
	result := ""
	links := session.SearchLinks(ownerId, srcModuleId, srcFieldId, dstModuleId)
	if len(links) > 0 {
		result = " AND ejaId IN (" + strings.Join(links, ",") + ") "
	}
	return result
}

func (session *TypeSession) AutoSearch(moduleId int64) (check bool) {
	hasSql, err := session.Value(`SELECT sqlCreated FROM ejaModules WHERE ejaId=?`, moduleId)
	if err == nil && session.Number(hasSql) > 0 {
		hasFields, err2 := session.Value(`SELECT COUNT(*) FROM ejaFields WHERE powerSearch > 0 AND ejaModuleId=?`, moduleId)
		if err2 == nil && session.Number(hasFields) == 0 {
			check = true
		}
	}
	return
}
