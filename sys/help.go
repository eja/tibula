// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"flag"
	"fmt"
)

func Help() {
	fmt.Println("Copyright:", "2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>")
	fmt.Println("Version:", Version)
	fmt.Println("Usage: tibula [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println()
}
