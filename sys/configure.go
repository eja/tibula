// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"flag"
)

func Configure() error {
	flag.BoolVar(&Commands.Start, "start", false, "start the web service")
	flag.BoolVar(&Commands.DbSetup, "db-setup", false, "initialize the database")
	flag.BoolVar(&Commands.Wizard, "wizard", false, "guided setup")
	flag.BoolVar(&Commands.Help, "help", false, "show this message")

	flag.StringVar(&Options.DbType, "db-type", "sqlite", "database type: sqlite/mysql")
	flag.StringVar(&Options.DbName, "db-name", "", "database name or filename")
	flag.StringVar(&Options.DbUser, "db-user", "", "database username")
	flag.StringVar(&Options.DbPass, "db-pass", "", "database password")
	flag.StringVar(&Options.DbHost, "db-host", "", "database hostname")
	flag.IntVar(&Options.DbPort, "db-port", 3306, "database port")
	flag.StringVar(&Options.DbSetupPath, "db-setup-path", "", "setup files path")
	flag.StringVar(&Options.DbSetupUser, "db-setup-user", "admin", "setup admin username")
	flag.StringVar(&Options.DbSetupPass, "db-setup-pass", "", "setup admin password")
	flag.StringVar(&Options.WebPath, "web-path", "", "web path")
	flag.StringVar(&Options.WebHost, "web-host", "localhost", "web listen address")
	flag.IntVar(&Options.WebPort, "web-port", 35248, "web listen port")
	flag.StringVar(&Options.WebTlsPublic, "web-tls-public", "", "web ssl/tls public certificate")
	flag.StringVar(&Options.WebTlsPrivate, "web-tls-private", "", "web ssl/tls private certificate")
	flag.StringVar(&Options.ConfigFile, "config", "", "json config file")
	flag.StringVar(&Options.Language, "language", "en", "default language code")
	flag.IntVar(&Options.LogLevel, "log-level", 3, "set the log level (1-5): 1=Error, 2=Warn, 3=Info, 4=Debug, 5=Trace")
	flag.Parse()

	if Options.ConfigFile != "" {
		if err := ConfigRead(Options.ConfigFile, &Options); err != nil {
			return err
		}
	}

	if Commands.DbSetup {
		if err := Setup(); err != nil {
			return err
		}
	}

	if Commands.Wizard {
		if err := wizardSetup(); err != nil {
			return err
		}
	}

	if Commands.Help {
		Help()
	}
	return nil
}
