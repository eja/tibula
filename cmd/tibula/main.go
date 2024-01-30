// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package main

import (
	"github.com/eja/tibula/sys"
	"github.com/eja/tibula/web"
	"log"
)

func main() {
	if err := sys.Configure(); err != nil {
		log.Fatal(err)
	}
	if sys.Commands.Start {
		if sys.Options.DbName == "" && sys.Options.ConfigFile == "" {
			if err := sys.ConfigRead("tibula.json", &sys.Options); err != nil {
				log.Fatal("Config file missing or not enough parameters to continue.")
			}
		}
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
