// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package api

import (
	"github.com/eja/tibula/db"
)

type TypeDbLink = db.TypeLink
type TypeDbGroup = db.TypeGroup
type TypeDbModule = db.TypeModule
type TypeDbSession = db.TypeSession

var DbSession = db.Session
