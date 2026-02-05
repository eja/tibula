// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import "github.com/eja/tibula/db"

type (
	DbLink    = db.TypeLink
	DbGroup   = db.TypeGroup
	DbModule  = db.TypeModule
	DbSession = db.TypeSession
)

var DbProvider = db.Session
