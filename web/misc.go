// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"regexp"
)

func arrayKeyNameExtract(input string) string {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(input)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}
