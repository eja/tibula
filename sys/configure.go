// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"os"
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
	flag.StringVar(&Options.LogFile, "log-file", "", "log file")
	flag.IntVar(&Options.LogLevel, "log-level", 3, "Detail level: 0=None, 1=Error, 2=Warn, 3=Info, 4=Debug")
	flag.StringVar(&Options.GoogleSsoId, "google-sso-id", "", "google sso client id")
	flag.Parse()

	parse := false

	if Options.ConfigFile != "" {
		if err := ConfigRead(Options.ConfigFile, &Options); err != nil {
			return err
		}
		parse = true
	} else {
		if err := ConfigRead(ConfigFileName(), &Options); err == nil {
			parse = true
		}
	}
	if parse {
		flag.Parse()
	}

	var w io.Writer = io.Discard
	var level slog.Level

	if Options.LogLevel > 0 {
		w = os.Stderr
		if Options.LogFile != "" {
			if f, err := os.OpenFile(Options.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err != nil {
				log.Fatal("Failed to open log file")
			} else {
				w = f
			}
		}

		switch {
		case Options.LogLevel == 1:
			level = slog.LevelError
		case Options.LogLevel == 2:
			level = slog.LevelWarn
		case Options.LogLevel == 3:
			level = slog.LevelInfo
		default:
			level = slog.LevelDebug
		}
	}

	opts := &slog.HandlerOptions{Level: level}
	logger := slog.New(slog.NewJSONHandler(w, opts))
	slog.SetDefault(logger)

	return nil
}
