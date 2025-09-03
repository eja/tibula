// Copyright (C) 2007-2025 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"encoding/csv"
	"regexp"
	"strings"
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
