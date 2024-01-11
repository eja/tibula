// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"fmt"
	"github.com/eja/tibula/sys"
	"log"
)

func info(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	*array = append(*array, row)
	if sys.Options.LogLevel > 3 {
		log.Println("[info]", row)
	}
}

func alert(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	*array = append(*array, row)
	if sys.Options.LogLevel > 3 {
		log.Println("[alert]", row)
	}
}
