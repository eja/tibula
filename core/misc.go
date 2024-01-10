// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"fmt"
	"log"
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
