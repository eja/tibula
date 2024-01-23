// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"embed"
	"errors"
	"fmt"
	"github.com/eja/tibula/sys"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed assets
var assets embed.FS

var Router = http.NewServeMux()
var RouterPathCore = "/"
var RouterPathStatic = "/static/"

func Start() error {
	address := fmt.Sprintf("%s:%d", sys.Options.WebHost, sys.Options.WebPort)

	Router.HandleFunc(RouterPathCore, Core)

	if sys.Options.WebPath != "" {
		staticDir := http.Dir(filepath.Join(sys.Options.WebPath, "static"))
		Router.Handle(RouterPathStatic, http.StripPrefix(RouterPathStatic, http.FileServer(staticDir)))
	} else {
		staticFs, err := fs.Sub(assets, "assets/static")
		if err != nil {
			return err
		}
		Router.Handle(RouterPathStatic, http.StripPrefix(RouterPathStatic, http.FileServer(http.FS(staticFs))))
	}

	if sys.Options.WebTlsPrivate != "" && sys.Options.WebTlsPublic != "" {
		if _, err := os.Stat(sys.Options.WebTlsPrivate); err != nil {
			return errors.New("failed to open private certificate")
		} else if _, err := os.Stat(sys.Options.WebTlsPublic); err != nil {
			return errors.New("failed to open public certificate")
		} else {
			log.Printf("Starting server on https://%s\n", address)
			if err := http.ListenAndServeTLS(address, sys.Options.WebTlsPublic, sys.Options.WebTlsPrivate, Router); err != nil {
				return err
			}
		}
	} else {
		log.Printf("Starting server on http://%s\n", address)
		if err := http.ListenAndServe(address, Router); err != nil {
			return err
		}
	}

	return nil
}
