// Copyright (C) 2007-2025 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"encoding/csv"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/eja/tibula/api"
)

func arrayKeyNameExtract(input string) string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(input)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func csvContains(csvData, searchString string) bool {
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return false
	}

	for _, record := range records {
		for _, field := range record {
			if field == searchString {
				return true
			}
		}
	}

	return false
}

func arrayToCsvQuoted(data []string) string {
	jsonBytes, _ := json.Marshal(data)
	jsonString := string(jsonBytes)

	return strings.Trim(jsonString, "[]")
}

func subModulePathExtract(value string) (subModulePath []api.SubModulePathItem) {
	for _, part := range strings.Split(value, ",") {
		pair := strings.Split(part, ".")
		if len(pair) == 3 {
			linkingModuleId, _ := strconv.Atoi(pair[0])
			moduleId, _ := strconv.Atoi(pair[1])
			fieldId, _ := strconv.Atoi(pair[2])
			subModulePath = append(subModulePath, api.SubModulePathItem{
				LinkingModuleId: int64(linkingModuleId),
				ModuleId:        int64(moduleId),
				FieldId:         int64(fieldId),
			})
		}
	}
	return
}
