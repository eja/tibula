// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"fmt"
	"log"
	"regexp"
)

func info(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	log.Println("[info]", row)
	*array = append(*array, row)
}

func alert(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	log.Println("[alert]", row)
	*array = append(*array, row)
}

func arrayKeyNameExtract(input string) string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(input)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}
