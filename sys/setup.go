// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"github.com/eja/tibula/db"
)

func Setup() error {
	if err := db.Open(Options.DbType, Options.DbName, Options.DbUser, Options.DbPass, Options.DbHost, Options.DbPort); err != nil {
		return err
	}
	if err := db.Setup(Options.DbSetupPath, Options.DbSetupUser, Options.DbSetupPass); err != nil {
		return err
	}
	return nil
}
