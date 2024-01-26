// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func wizardPrompt(message string) string {
	fmt.Printf("%s: ", message)
	var input string
	fmt.Scanln(&input)
	return input
}

func wizardPassword(message string) string {
	fmt.Printf("%s: ", message)
	pass, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return strings.TrimSpace(string(pass))
}

func wizardSetup() error {
	var cwd = filepath.Dir(os.Args[0])
	var tibulaJson = filepath.Join(cwd, "tibula.json")
	var tibulaDb = filepath.Join(cwd, "tibula.db")

	if _, err := os.Stat(tibulaJson); err == nil {
		return errors.New("A configuration file already exists, remove it and try again.")
	}

	//admin
	setupUser := wizardPrompt("Choose an administrator username (admin)")
	if setupUser != "" {
		Options.SetupUser = setupUser
	}
	Options.SetupPass = wizardPassword("Enter the administrator password")
	if Options.SetupPass == "" {
		return errors.New("Password cannot be empty.")
	}
	passConfirm := wizardPassword("Repeat password")
	if passConfirm != Options.SetupPass {
		return errors.New("Passwords do not match.")
	}
	//web
	webHost := wizardPrompt("Web address to listen to (localhost)")
	if webHost != "" {
		Options.WebHost = webHost
	}
	webPort := wizardPrompt("Web port to listen to (35248)")
	if webPort != "" {
		Options.WebPort, _ = strconv.Atoi(webPort)
	}
	webPrivate := wizardPrompt("Web https private certificate path (none)")
	if webPrivate != "" {
		Options.WebTlsPrivate = webPrivate
		webPublic := wizardPrompt("Web https public certificate path (none)")
		if webPublic != "" {
			Options.WebTlsPublic = webPublic
		}
	}
	//db
	dbType := wizardPrompt("Choose a database engine between sqlite and mysql (sqlite)")
	if dbType == "mysql" {
		Options.DbType = dbType
		Options.DbName = wizardPrompt("Database name")
		Options.DbUser = wizardPrompt("Database username")
		Options.DbPass = wizardPassword("Database password")
		dbHost := wizardPrompt("Database hostname (localhost)")
		if dbHost != "" {
			Options.DbHost = dbHost
		}
		dbPort := wizardPrompt("Database port (3306)")
		if dbPort != "" {
			Options.DbPort, _ = strconv.Atoi(dbPort)
		}
	} else {
		Options.DbPort = 0
		dbName := wizardPrompt(fmt.Sprintf("Database file name (%s)", tibulaDb))
		if dbName == "" {
			Options.DbName = tibulaDb
		} else {
			Options.DbName = dbName
		}
		if _, err := os.Stat(Options.DbName); err == nil {
			return errors.New("An sqlite database with the same name already exists, remove it and try again.")
		}
	}
	//misc
	language := wizardPrompt("Choose default language (en)")
	if language != "" {
		Options.Language = language
	}
	logLevel := wizardPrompt("Choose log level between 1=Error, 2=Warn, 3=Info, 4=Debug, 5=Trace (3)")
	if logLevel != "" {
		Options.LogLevel, _ = strconv.Atoi(logLevel)
	}
	jsonFile := wizardPrompt(fmt.Sprintf("Config file (%s)", tibulaJson))
	if jsonFile == "" {
		jsonFile = tibulaJson
	}

	if err := Setup(); err != nil {
		return err
	}

	Options.ConfigFile = ""
	Options.SetupUser = ""
	Options.SetupPass = ""
	Options.Setup = false
	if err := ConfigWrite(jsonFile); err != nil {
		return fmt.Errorf("Cannot write the configuration file, %w\n", err)
	}

	return nil
}
