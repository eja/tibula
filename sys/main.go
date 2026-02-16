// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"flag"
	"fmt"
)

const Name = "tibula"
const Label = "Tibula"
const Version = "19.2.16"

var Options TypeConfig
var Commands TypeCommand

func Help() {
	fmt.Println("Copyright:", "2007-2026 by Ubaldo Porcheddu <ubaldo@eja.it>")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Usage: %s [options]\n", Name)
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println()
}
