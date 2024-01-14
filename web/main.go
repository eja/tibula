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

func Start() error {
	mux := http.NewServeMux()
	address := fmt.Sprintf("%s:%d", sys.Options.WebHost, sys.Options.WebPort)

	mux.HandleFunc("/", home)

	if sys.Options.WebPath != "" {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(sys.Options.WebPath, "static")))))
	} else {
		staticFs, err := fs.Sub(assets, "assets/static")
		if err != nil {
			return err
		}
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFs))))
	}

	if sys.Options.WebTlsPrivate != "" && sys.Options.WebTlsPublic != "" {
		if _, err := os.Stat(sys.Options.WebTlsPrivate); err != nil {
			return errors.New("cannot open private certificate")
		} else {
			if _, err := os.Stat(sys.Options.WebTlsPublic); err != nil {
				return errors.New("cannot open public certificate")
			} else {
				log.Printf("Starting server on https://%s\n", address)
				if err := http.ListenAndServeTLS(address, sys.Options.WebTlsPublic, sys.Options.WebTlsPrivate, mux); err != nil {
					return err
				}
			}
		}
	} else {
		log.Printf("Starting server on http://%s\n", address)
		if err := http.ListenAndServe(address, mux); err != nil {
			return err
		}
	}

	return nil
}
