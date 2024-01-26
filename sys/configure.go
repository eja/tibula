// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func Configure() {
	help := flag.Bool("help", false, "show this message")
	wizard := flag.Bool("wizard", false, "guided setup")
	flag.StringVar(&Options.DbType, "db-type", "sqlite", "database type: sqlite/mysql")
	flag.StringVar(&Options.DbName, "db-name", "", "database name or filename")
	flag.StringVar(&Options.DbUser, "db-user", "", "database username")
	flag.StringVar(&Options.DbPass, "db-pass", "", "database password")
	flag.StringVar(&Options.DbHost, "db-host", "", "database hostname")
	flag.IntVar(&Options.DbPort, "db-port", 3306, "database port")
	flag.StringVar(&Options.WebPath, "web-path", "", "web path")
	flag.StringVar(&Options.WebHost, "web-host", "localhost", "web listen address")
	flag.IntVar(&Options.WebPort, "web-port", 35248, "web listen port")
	flag.StringVar(&Options.WebTlsPublic, "web-tls-public", "", "web ssl/tls public certificate")
	flag.StringVar(&Options.WebTlsPrivate, "web-tls-private", "", "web ssl/tls private certificate")
	flag.BoolVar(&Options.Setup, "setup", false, "initialize the database")
	flag.StringVar(&Options.SetupPath, "setup-path", "", "setup files path")
	flag.StringVar(&Options.SetupUser, "setup-user", "admin", "setup admin username")
	flag.StringVar(&Options.SetupPass, "setup-pass", "", "setup admin password")
	flag.BoolVar(&Options.Start, "start", false, "start the web service")
	flag.StringVar(&Options.ConfigFile, "config", "", "json config file")
	flag.StringVar(&Options.Language, "language", "en", "default language code")
	flag.IntVar(&Options.LogLevel, "log-level", 3, "set the log level (1-5): 1=Error, 2=Warn, 3=Info, 4=Debug, 5=Trace")
	flag.Parse()

	if Options.ConfigFile != "" {
		if err := ConfigRead(Options.ConfigFile); err != nil {
			log.Fatalf("Cannot open config file: %v\n", err)
		}
	}

	if *wizard {
		if err := wizardSetup(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *help {
		Help()
		os.Exit(0)
	}
}
