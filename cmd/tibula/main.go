// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package main

import (
	"github.com/eja/tibula/log"
	"github.com/eja/tibula/sys"
	"github.com/eja/tibula/web"
)

func main() {
	if err := sys.Configure(); err != nil {
		log.Fatal(err)
	}

	if sys.Commands.DbSetup {
		if err := sys.Setup(); err != nil {
			log.Fatal(err)
		}
	} else if sys.Commands.Wizard {
		if err := sys.WizardSetup(); err != nil {
			log.Fatal(err)
		}
	} else if sys.Commands.Start {
		if sys.Options.DbName == "" {
			log.Fatal("Database name/file is mandatory.")
		}
		if err := web.Start(); err != nil {
			log.Fatal("Cannot start the web service: ", err)
		}
	} else {
		sys.Help()
	}
}
