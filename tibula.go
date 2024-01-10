// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package main

import (
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/sys"
	"github.com/eja/tibula/web"
	"log"
)

func main() {
	sys.Configure()

	if sys.Options.Setup {
		err := db.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort)
		if err != nil {
			log.Fatal(err)
		}
		if err := db.Setup(sys.Options.SetupPath, sys.Options.SetupUser, sys.Options.SetupPass); err != nil {
			log.Fatal(err)
		}
	} else if sys.Options.Start {
		if sys.Options.DbName == "" {
			log.Fatal("Database name/file is mandatory")
		}
		if err := web.Start(); err != nil {
			log.Fatal("Cannot start the web service: ", err)
		}
	} else {
		sys.Help()
	}
}
