// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"fmt"
	"github.com/eja/tibula/log"
	"github.com/eja/tibula/sys"
)

func info(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	*array = append(*array, row)
	if sys.Options.LogLevel > 3 {
		log.Trace("[api] [info]", row)
	}
}

func alert(array *[]string, format string, args ...interface{}) {
	row := fmt.Sprintf(format, args...)
	*array = append(*array, row)
	if sys.Options.LogLevel > 3 {
		log.Trace("[api] [alert]", row)
	}
}
