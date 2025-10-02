// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func (session TypeSession) Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (session *TypeSession) Password(value string) string {
	return session.Sha256(value)
}

func (session *TypeSession) Sha256(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func (session *TypeSession) NumbersToCsv(slice []int64) string {
	result := ""
	for i, v := range slice {
		result += fmt.Sprint(v)
		if i < len(slice)-1 {
			result += ","
		}
	}
	return result
}

func (session *TypeSession) SelectToRows(value string) []TypeSelect {
	var result []TypeSelect
	i := 0

	value = strings.ReplaceAll(value, "\r", "")

	if strings.Contains(value, "|") {
		re := regexp.MustCompile(`([^|\n]*)\|([^|\n]*)`)
		matches := re.FindAllStringSubmatch(value, -1)

		for _, match := range matches {
			row := TypeSelect{match[1], match[2]}
			result = append(result, row)
		}
	} else {
		rows := strings.Split(value, "\n")
		for _, row := range rows {
			rowData := TypeSelect{row, row}
			result = append(result, rowData)

			if i == 0 {
				i = 1
			} else {
				i = 0
			}
		}
	}

	return result
}

func (session *TypeSession) SelectSqlToRows(query string) []TypeSelect {
	var result []TypeSelect
	cols, err := session.Cols(query)
	if err == nil {
		rows, err := session.Rows(query)
		if err == nil {
			for _, row := range rows {
				result = append(result, TypeSelect{row[cols[0]], row[cols[1]]})
			}
		}
	}
	return result
}

func (session *TypeSession) IncludeList(query string, args ...interface{}) ([]int64, error) {
	response := make([]int64, 0)

	rows, err := session.Rows(query, args...)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		for _, value := range row {
			response = append(response, session.Number(value))
			break
		}
	}
	return response, nil
}
