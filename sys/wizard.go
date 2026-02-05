// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func WizardPrompt(message string) string {
	fmt.Printf("%s: ", message)
	var input string
	fmt.Scanln(&input)
	return input
}

func WizardPassword(message string) string {
	fmt.Printf("%s: ", message)
	pass, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return strings.TrimSpace(string(pass))
}

func WizardSetup() error {
	var tibulaDb = "tibula.db"
	var tibulaJson = ConfigFileName()

	if _, err := os.Stat(tibulaJson); err == nil {
		return errors.New("A configuration file already exists, remove it and try again.")
	}

	//admin
	setupUser := WizardPrompt("Choose an administrator username (admin)")
	if setupUser != "" {
		Options.DbSetupUser = setupUser
	}
	Options.DbSetupPass = WizardPassword("Enter the administrator password")
	if Options.DbSetupPass == "" {
		return errors.New("Password cannot be empty.")
	}
	passConfirm := WizardPassword("Repeat password")
	if passConfirm != Options.DbSetupPass {
		return errors.New("Passwords do not match.")
	}
	//web
	webHost := WizardPrompt("Web address to listen to (localhost)")
	if webHost != "" {
		Options.WebHost = webHost
	}
	webPort := WizardPrompt("Web port to listen to (35248)")
	if webPort != "" {
		Options.WebPort, _ = strconv.Atoi(webPort)
	}
	webPrivate := WizardPrompt("Web https private certificate path (none)")
	if webPrivate != "" {
		Options.WebTlsPrivate = webPrivate
		webPublic := WizardPrompt("Web https public certificate path (none)")
		if webPublic != "" {
			Options.WebTlsPublic = webPublic
		}
	}
	//db
	dbType := WizardPrompt("Choose a database engine between sqlite and mysql (sqlite)")
	if dbType == "mysql" {
		Options.DbType = dbType
		Options.DbName = WizardPrompt("Database name")
		Options.DbUser = WizardPrompt("Database username")
		Options.DbPass = WizardPassword("Database password")
		dbHost := WizardPrompt("Database hostname (localhost)")
		if dbHost != "" {
			Options.DbHost = dbHost
		}
		dbPort := WizardPrompt("Database port (3306)")
		if dbPort != "" {
			Options.DbPort, _ = strconv.Atoi(dbPort)
		}
	} else {
		Options.DbPort = 0
		dbName := WizardPrompt(fmt.Sprintf("Database file name (%s)", tibulaDb))
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
	language := WizardPrompt("Choose default language (en)")
	if language != "" {
		Options.Language = language
	}
	logLevel := WizardPrompt("Choose log level between 1=Error, 2=Warn, 3=Info, 4=Debug, 5=Trace (3)")
	if logLevel != "" {
		Options.LogLevel, _ = strconv.Atoi(logLevel)
	}
	Options.LogFile = WizardPrompt("Choose a log file (stderr)")
	jsonFile := WizardPrompt(fmt.Sprintf("Config file (%s)", tibulaJson))
	if jsonFile == "" {
		jsonFile = tibulaJson
	}

	if err := Setup(); err != nil {
		return err
	}

	Options.ConfigFile = ""
	Options.DbSetupUser = ""
	Options.DbSetupPass = ""
	if err := ConfigWrite(jsonFile, &Options); err != nil {
		return fmt.Errorf("Cannot write the configuration file, %w\n", err)
	}
	Options.ConfigFile = jsonFile

	return nil
}
