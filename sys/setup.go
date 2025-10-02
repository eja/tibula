// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"github.com/eja/tibula/db"
)

func Setup() (err error) {
	db := db.Session()
	if err = db.Open(Options.DbType, Options.DbName, Options.DbUser, Options.DbPass, Options.DbHost, Options.DbPort); err != nil {
		return
	}
	if err = db.Setup(Options.DbSetupPath); err != nil {
		return
	}
	if err = db.SetupAdmin(Options.DbSetupUser, Options.DbSetupPass); err != nil {
		return
	}

	return db.Close()
}
