// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (session TypeSession) Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (session *TypeSession) Password(value string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return session.Sha256(value)
	}
	return string(hash)
}

func (session *TypeSession) PasswordCheck(password string, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err == nil {
		return true
	}

	return session.Sha256(password) == storedHash
}

func (session *TypeSession) Sha256(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func (session *TypeSession) NumbersToCsv(slice []int64) string {
	var result strings.Builder
	for i, v := range slice {
		result.WriteString(fmt.Sprint(v))
		if i < len(slice)-1 {
			result.WriteString(",")
		}
	}
	return result.String()
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
		rows := strings.SplitSeq(value, "\n")
		for row := range rows {
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

func (session *TypeSession) IncludeList(query string, args ...any) ([]int64, error) {
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
