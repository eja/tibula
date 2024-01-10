// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package main

import (
	"log"
	"tibula/internal/cfg"
	"tibula/internal/db"
	"tibula/internal/web"
)

func main() {
	cfg.Configure()

	if cfg.Options.Setup {
		err := db.Open(cfg.Options.DbType, cfg.Options.DbName, cfg.Options.DbUser, cfg.Options.DbPass, cfg.Options.DbHost, cfg.Options.DbPort)
		if err != nil {
			log.Fatal(err)
		}
		if err := db.Setup(cfg.Options.SetupPath, cfg.Options.SetupUser, cfg.Options.SetupPass); err != nil {
			log.Fatal(err)
		}
	} else if cfg.Options.Start {
		if cfg.Options.DbName == "" {
			log.Fatal("Database name/file is mandatory")
		}
		if err := web.Start(); err != nil {
			log.Fatal("Cannot start the web service: ", err)
		}
	} else {
		cfg.Help()
	}
}
